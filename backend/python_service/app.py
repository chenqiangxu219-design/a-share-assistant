import json
import logging
import time
from contextlib import asynccontextmanager

import akshare as ak
import pandas as pd
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from tdxpy.hq import TdxHq_API

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# tdxpy connection (singleton, lazy init with server config)
_tdx_api = None
_tdx_connected = False
TDX_HOST = "202.108.253.139"
TDX_PORT = 7709


def get_tdx_api():
    global _tdx_api, _tdx_connected
    if _tdx_api is None:
        _tdx_api = TdxHq_API()
    if not _tdx_connected:
        try:
            _tdx_api.connect(TDX_HOST, TDX_PORT)
        except Exception:
            _tdx_api.connect()
        _tdx_connected = True
    return _tdx_api


# Cache for quote/kline data (shorter TTL during trading hours)
_quote_cache = {}
_quote_cache_time = {}
_kline_cache = {}
_kline_cache_time = {}
QUOTE_CACHE_TTL = 60   # 1 minute during trading
KLINES_CACHE_TTL = 300  # 5 minutes for historical


def tdx_get_quote(code: str):
    """Get quote via tdxpy. Returns dict or None if off-hours/error."""
    try:
        api = get_tdx_api()
        market_id = 1 if code.startswith("6") else 0
        result = api.get_security_quotes(code=[(market_id, code)])
        if not result or len(result) == 0:
            return None
        item = result[0]
        # TDX quote fields: market, code, price, open, high, low, vol, bid1-5, ask1-5, ...
        return {
            "code": code,
            "name": "",
            "price": float(item.get("price", 0) if isinstance(item, dict) else 0),
            "open": float(item.get("open", 0) if isinstance(item, dict) else 0),
            "high": float(item.get("high", 0) if isinstance(item, dict) else 0),
            "low": float(item.get("low", 0) if isinstance(item, dict) else 0),
            "yesterday_close": float(item.get("pre_close", 0) if isinstance(item, dict) else 0),
            "volume": float(item.get("volume", 0) if isinstance(item, dict) else 0),
            "amount": float(item.get("amount", 0) if isinstance(item, dict) else 0),
            "bid_prices": [],
            "bid_volumes": [],
            "ask_prices": [],
            "ask_volumes": [],
        }
    except Exception as e:
        logger.warning(f"tdxpy quote failed: {e}")
        return None


def tdx_get_klines(code: str, period: str, count: int):
    """Get K-lines via tdxpy. Returns list or empty if off-hours/error."""
    try:
        api = get_tdx_api()
        market_id = 1 if code.startswith("6") else 0
        period_map = {"d": 0, "w": 1, "m": 2, "60": 4, "30": 5, "15": 6, "5": 7, "1": 8}
        cat = period_map.get(period, 0)
        df = api.get_security_bars(cat, market_id, code, count - 1, count)
        if df is None or df.empty:
            return []
        rows = []
        for _, row in df.iterrows():
            rows.append({
                "date": str(row.get("date", "")),
                "open": float(row.get("open", 0)),
                "high": float(row.get("high", 0)),
                "low": float(row.get("low", 0)),
                "close": float(row.get("close", 0)),
                "volume": float(row.get("volume", 0)),
            })
        return rows
    except Exception as e:
        logger.warning(f"tdxpy kline failed: {e}")
        return []


def ak_get_quote(code: str):
    """Get quote via akshare stock_zh_a_hist (last day's data as fallback)."""
    try:
        df = ak.stock_zh_a_hist(symbol=code, period="daily", adjust="qfq", count=1)
        if df is None or df.empty:
            return None
        row = df.iloc[0]
        return {
            "code": code,
            "name": "",
            "price": float(row.get("收盘", 0)),
            "open": float(row.get("开盘", 0)),
            "high": float(row.get("最高", 0)),
            "low": float(row.get("最低", 0)),
            "yesterday_close": float(row.get("昨收", 0) if "昨收" in row.index else 0),
            "volume": float(row.get("成交量", 0)),
            "amount": float(row.get("成交额", 0)),
            "bid_prices": [],
            "bid_volumes": [],
            "ask_prices": [],
            "ask_volumes": [],
        }
    except Exception as e:
        logger.warning(f"akshare quote failed: {e}")
        return None


def ak_get_klines(code: str, period: str, count: int):
    """Get K-lines via akshare stock_zh_a_hist."""
    try:
        period_map = {"d": "daily", "w": "weekly", "m": "monthly", "60": "60", "30": "30", "15": "15", "5": "5", "1": "1"}
        p = period_map.get(period, "daily")
        df = ak.stock_zh_a_hist(symbol=code, period=p, adjust="qfq", count=count)
        if df is None or df.empty:
            return []
        rows = []
        for _, row in df.iterrows():
            rows.append({
                "date": str(row.get("日期", "")),
                "open": float(row.get("开盘", 0)),
                "high": float(row.get("最高", 0)),
                "low": float(row.get("最低", 0)),
                "close": float(row.get("收盘", 0)),
                "volume": float(row.get("成交量", 0)),
            })
        return rows
    except Exception as e:
        logger.warning(f"akshare kline failed: {e}")
        return []


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Python microservice starting")
    yield
    logger.info("Python microservice shutting down")


app = FastAPI(title="A-Share Data Service", lifespan=lifespan)


# --- quote/kline endpoints (hybrid: tdxpy → akshare) ---

@app.get("/mootdx/quote/{code}")
def mootdx_quote(code: str):
    """Get real-time quote. Tries tdxpy first, falls back to akshare."""
    global _quote_cache, _quote_cache_time
    now = time.time()
    cached_time = _quote_cache_time.get(code, 0)
    if code in _quote_cache and (now - cached_time) < QUOTE_CACHE_TTL:
        return _quote_cache[code]

    result = tdx_get_quote(code)
    if result is None:
        result = ak_get_quote(code)
    if result is None:
        raise HTTPException(404, f"Stock {code} not found")

    _quote_cache[code] = result
    _quote_cache_time[code] = now
    return result


@app.get("/mootdx/kline/{code}")
def mootdx_kline(code: str, period: str = "d", count: int = 100):
    """Get K-line data. Tries tdxpy first, falls back to akshare."""
    global _kline_cache, _kline_cache_time
    cache_key = f"{code}_{period}_{count}"
    now = time.time()
    cached_time = _kline_cache_time.get(cache_key, 0)
    if cache_key in _kline_cache and (now - cached_time) < KLINES_CACHE_TTL:
        return _kline_cache[cache_key]

    result = tdx_get_klines(code, period, count)
    if not result:
        result = ak_get_klines(code, period, count)

    _kline_cache[cache_key] = result
    _kline_cache_time[cache_key] = now
    return result


# --- akshare endpoints ---

def akshare_request(func, *args, **kwargs):
    """Wrap akshare call with retry and rate-limit handling."""
    for attempt in range(3):
        try:
            return func(*args, **kwargs)
        except Exception as e:
            if attempt < 2:
                logger.warning(f"akshare request failed (attempt {attempt+1}): {e}. Retrying...")
                time.sleep(2 * (attempt + 1))  # exponential backoff
            else:
                raise

@app.get("/akshare/news/{code}")
def akshare_news(code: str):
    """Get stock news via akshare."""
    try:
        df = akshare_request(ak.stock_news_em, symbol=code)
        if df is None or df.empty:
            return []
        result = []
        for _, row in df.head(50).iterrows():
            item = {
                "title": str(row.get("新闻标题", "")),
                "digest": str(row.get("新闻内容", "")),
                "info_content": str(row.get("新闻内容", "")),
                "info_url": str(row.get("新闻链接", "")),
                "showtime": str(row.get("发布时间", "")),
            }
            result.append(item)
        return result
    except Exception as e:
        logger.error(f"akshare news error: {e}")
        raise HTTPException(500, str(e))


@app.get("/akshare/research/{code}")
def akshare_research(code: str):
    """Get research reports related to a stock."""
    try:
        df = akshare_request(ak.stock_research_report_em, symbol=code)
        if df is None or df.empty:
            return []
        result = []
        for _, row in df.head(30).iterrows():
            item = {
                "title": str(row.get("报告名称", "")),
                "author": str(row.get("机构", "")),
                "date": str(row.get("日期", "")),
                "digest": str(row.get("东财评级", "")),
                "url": str(row.get("报告PDF链接", "")),
            }
            result.append(item)
        return result
    except Exception as e:
        logger.error(f"akshare research error: {e}")
        raise HTTPException(500, str(e))


@app.get("/akshare/heatmap")
def akshare_heatmap():
    """Get sector/industry heatmap data. Uses THS source (more stable). Cached for 5 minutes."""
    global _heatmap_cache, _heatmap_cache_time
    now = time.time()
    if _heatmap_cache and (now - _heatmap_cache_time) < CACHE_TTL:
        return _heatmap_cache
    try:
        df = akshare_request(ak.stock_board_industry_summary_ths)
        if df is None or df.empty:
            return []
        result = []
        for _, row in df.iterrows():
            item = {
                "name": str(row.get("板块", "")),
                "change_pct": float(row.get("涨跌幅", 0)),
                "lead_stock": str(row.get("领涨股", "")),
                "lead_stock_pct": float(row.get("领涨股-涨跌幅", 0)),
            }
            result.append(item)
        _heatmap_cache = result
        _heatmap_cache_time = now
        return result
    except Exception as e:
        logger.error(f"akshare heatmap error: {e}")
        if _heatmap_cache:
            return _heatmap_cache  # return stale data
        raise HTTPException(500, str(e))


# --- Cached hot stocks (Xueqiu source, ~13s fetch, cache 5min) ---

_hot_stocks_cache = []
_hot_stocks_cache_time = 0
_heatmap_cache = []
_heatmap_cache_time = 0
CACHE_TTL = 300  # 5 minutes


def fetch_hot_stocks_from_xueqiu():
    """Fetch hot stocks from Xueqiu via akshare."""
    global _hot_stocks_cache, _hot_stocks_cache_time
    try:
        df = ak.stock_hot_follow_xq()
        if df is None or df.empty:
            return []
        result = []
        for _, row in df.head(30).iterrows():
            code_raw = str(row.get("股票代码", ""))
            code = code_raw.replace("SH", "").replace("SZ", "")
            result.append({
                "code": code,
                "name": str(row.get("股票简称", "")),
                "change_pct": 0,  # Xueqiu doesn't provide change_pct directly
                "volume": float(row.get("关注", 0)),  # follower count as热度
                "concept": "",
            })
        _hot_stocks_cache = result
        _hot_stocks_cache_time = time.time()
        return result
    except Exception as e:
        logger.error(f"fetch hot stocks from xueqiu error: {e}")
        return []


@app.get("/akshare/hot-stocks")
def akshare_hot_stocks():
    """Get hot trending stocks. Uses Xueqiu with 5min cache."""
    now = time.time()
    # Return cached data if not expired
    if _hot_stocks_cache and (now - _hot_stocks_cache_time) < CACHE_TTL:
        return _hot_stocks_cache
    # Cache miss or expired — fetch in background, return stale data if available
    if (now - _hot_stocks_cache_time) >= CACHE_TTL:
        # Trigger refresh (will block, but only after TTL)
        try:
            data = fetch_hot_stocks_from_xueqiu()
            if data:
                return data
        except Exception as e:
            logger.warning(f"hot stocks refresh failed: {e}")
    # Return stale cache
    if _hot_stocks_cache:
        return _hot_stocks_cache
    # First call ever — fetch synchronously
    data = fetch_hot_stocks_from_xueqiu()
    return data if data else []


@app.get("/akshare/sector-boards")
def akshare_sector_boards():
    """Get concept boards. Uses stock_board_change_em (more stable than concept_name_em)."""
    try:
        df = akshare_request(ak.stock_board_change_em)
        if df is None or df.empty:
            return []
        result = []
        for _, row in df.head(80).iterrows():
            item = {
                "name": str(row.get("板块名称", "")),
                "change_pct": float(row.get("涨跌幅", 0)),
            }
            result.append(item)
        return result
    except Exception as e:
        logger.error(f"akshare sector boards error: {e}")
        raise HTTPException(500, str(e))


# --- Web feed (Hacker News source) ---

_web_feed_cache = []
_web_feed_cache_time = 0


def _extract_domain(url: str) -> str:
    """Extract domain from URL."""
    if not url or url.startswith("item/"):
        return ""
    try:
        from urllib.parse import urlparse
        hostname = urlparse(url).hostname
        return hostname or ""
    except Exception:
        return ""


def fetch_web_feed(count: int = 50) -> list[dict]:
    """Fetch top stories from Hacker News."""
    import requests

    global _web_feed_cache, _web_feed_cache_time
    try:
        story_ids = requests.get(
            "https://hacker-news.firebaseio.com/v0/topstories.json"
        ).json()
        stories = []
        for sid in story_ids[:count]:
            raw = requests.get(
                f"https://hacker-news.firebaseio.com/v0/item/{sid}.json"
            ).json()
            stories.append({
                "id": sid,
                "title": raw.get("title", ""),
                "url": raw.get("url", ""),
                "domain": _extract_domain(raw.get("url", "")),
                "score": raw.get("score", 0),
                "timestamp": raw.get("time", 0),
            })
            time.sleep(0.1)  # rate limit
        _web_feed_cache = stories
        _web_feed_cache_time = time.time()
        return stories
    except Exception as e:
        logger.error(f"web feed fetch error: {e}")
        return []


@app.get("/web/feed")
def web_feed(count: int = 50):
    """Get top stories from web news feed (Hacker News). Cached for 5 minutes."""
    global _web_feed_cache, _web_feed_cache_time
    now = time.time()
    if _web_feed_cache and (now - _web_feed_cache_time) < CACHE_TTL:
        return _web_feed_cache[:count]
    # Cache expired — fetch fresh data
    stories = fetch_web_feed(count)
    return stories[:count] if stories else _web_feed_cache[:count]


@app.get("/web/analysis")
def web_analysis():
    """Get analysis stats for the web feed."""
    stories = _web_feed_cache or fetch_web_feed(50)
    if not stories:
        return {"total_stories": 0}
    df = pd.DataFrame(stories)
    return {
        "total_stories": len(stories),
        "score_stats": {
            "mean": round(float(df["score"].mean()), 1),
            "median": round(float(df["score"].median()), 1),
            "max": int(df["score"].max()),
        },
        "top_domains": df[df["domain"] != ""]["domain"].value_counts().head(10).to_dict(),
        "hot_keywords": _extract_hot_keywords(stories),
    }


def _extract_hot_keywords(stories: list[dict]) -> dict:
    """Extract top keywords from story titles."""
    stop_words = {"the", "a", "an", "to", "for", "of", "in", "with", "my", "is", "on", "it", "as"}
    all_words = []
    for story in stories:
        words = str(story.get("title", "")).lower().split()
        all_words.extend([w.strip(".,:;!?\"'") for w in words if w not in stop_words])
    return pd.Series(all_words).value_counts().head(15).to_dict()


# --- health ---

@app.get("/health")
def health():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)
