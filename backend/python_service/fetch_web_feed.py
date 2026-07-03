#!/usr/bin/env python3
"""Dual-mode web feed pipeline: General (HN) + Targeted (akshare + Bing)."""

import argparse
import json
import time
from urllib.parse import urlparse

import pandas as pd
import requests


# --- Shared helpers ---

def _extract_domain(url: str) -> str:
    """Extract domain from URL."""
    if not url or url.startswith("item/"):
        return ""
    try:
        hostname = urlparse(url).hostname
        return hostname or ""
    except Exception:
        return ""


def analyze_feed(stories: list[dict]) -> dict:
    """Analyze the feed for patterns."""
    df = pd.DataFrame(stories)

    score_stats = {
        "mean": round(float(df["score"].mean()), 1),
        "median": round(float(df["score"].median()), 1),
        "max": int(df["score"].max()),
        "min": int(df["score"].min()),
    }

    domain_counts = df[df["domain"] != ""]["domain"].value_counts().head(10)

    stop_words = {"the", "a", "an", "to", "for", "of", "in", "with", "my", "is", "on", "it", "as"}
    all_words = []
    for title in df["title"]:
        words = str(title).lower().split()
        all_words.extend([w.strip(".,:;!?\"'") for w in words if w not in stop_words])
    word_freq = pd.Series(all_words).value_counts().head(15)

    return {
        "total_stories": len(stories),
        "score_stats": score_stats,
        "top_domains": domain_counts.to_dict(),
        "hot_keywords": word_freq.to_dict(),
    }


# --- Mode B: General HN Feed (Firebase — works from China) ---

def fetch_hn_top_stories(count: int = 50) -> list[dict]:
    """Fetch top stories from Hacker News API."""
    story_ids = requests.get(
        "https://hacker-news.firebaseio.com/v0/topstories.json"
    ).json()
    stories = []
    for sid in story_ids[:count]:
        try:
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
                "source": "hacker-news",
            })
            time.sleep(0.1)
        except Exception as e:
            print(f"Failed to fetch story {sid}: {e}")
            continue
    return stories


# --- Mode A: Targeted Search (Bing HTML + akshare news) ---

def fetch_bing_search(query: str, count: int = 20) -> list[dict]:
    """Search Bing for targeted results.

    Note: from China, Bing returns a rich page with knowledge panels,
    "People Also Ask", and image carousels. We extract the organic links
    by looking for <a> tags with substantial text content.
    """
    stories = []
    try:
        resp = requests.get(
            "https://www.bing.com/search",
            params={"q": query},
            headers={
                "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
                              "AppleWebKit/537.36 (KHTML, like Gecko) "
                              "Chrome/120.0.0.0 Safari/537.36",
                "Accept-Language": "en-US,en;q=0.9",
            },
            timeout=10,
        )
        if resp.status_code != 200:
            return []

        from bs4 import BeautifulSoup
        soup = BeautifulSoup(resp.text, "html.parser")

        # Strategy 1: Look for <li class="b_algo"> (classic Bing structure)
        for algo in soup.find_all("li", class_="b_algo"):
            title_el = algo.find("h2")
            link_el = algo.find("a", href=True)
            if title_el and link_el:
                stories.append({
                    "title": title_el.get_text(),
                    "url": link_el["href"],
                    "source": "bing",
                })

        # Strategy 2: If b_algo is empty, scan all <a> tags for "title-like" text
        if not stories:
            seen_urls = set()
            for a in soup.find_all("a", href=True):
                text = a.get_text().strip()
                url = a["href"]
                # Filter: substantial text, looks like a title, not a tracking link
                if (len(text) > 15 and
                        text.endswith((".", "s", "s's")) and
                        url.startswith("http") and
                        "/rp.aspx" not in url and
                        url not in seen_urls):
                    seen_urls.add(url)
                    stories.append({
                        "title": text,
                        "url": url,
                        "source": "bing",
                    })

        print(f"Bing returned {len(stories)} results")
    except Exception as e:
        print(f"Bing search failed: {e}")

    # Normalize to shared schema
    for s in stories:
        s.update({
            "id": len(stories),  # placeholder ID
            "domain": _extract_domain(s.get("url", "")),
            "score": 0,
            "timestamp": int(time.time()),
        })

    return stories[:count]


def fetch_akshare_web(query: str, count: int = 20) -> list[dict]:
    """Use akshare to fetch news (Chinese-network-friendly).

    Uses ak.stock_news_em which pulls from East Money.
    Note: this function usually takes a stock code, but we can use it
    for general "tech news" by passing common tech-stock codes.
    """
    import akshare as ak

    stories = []
    # For a general "web news" feel, pull from a few tech-heavy stocks
    # and filter by keyword. This is a pragmatic China-network approach.
    tech_codes = ["002230", "600519", "000001"]  # YYW, Moutai, Index
    for code in tech_codes:
        try:
            df = ak.stock_news_em(symbol=code)
            if df is None or df.empty:
                continue
            for _, row in df.head(10).iterrows():
                title = str(row.get("新闻标题", ""))
                # Simple keyword filter: check if query words appear in title
                if any(word in title for word in query.split()):
                    stories.append({
                        "id": len(stories) + 1,
                        "title": title,
                        "url": str(row.get("新闻链接", "")),
                        "domain": _extract_domain(str(row.get("新闻链接", ""))),
                        "score": 0,
                        "timestamp": int(time.time()),
                        "source": "akshare",
                    })
        except Exception as e:
            print(f"akshare news for {code}: {e}")
            continue

    return stories[:count]


# --- Main orchestrator ---

def fetch(mode: str = "general", query: str = "", count: int = 50) -> list[dict]:
    """Fetch stories based on mode."""
    if mode == "general":
        print(f"Fetching {count} HN top stories...")
        return fetch_hn_top_stories(count)
    elif mode == "targeted":
        if not query:
            print("Targeted mode requires a --query")
            return []
        print(f"Searching for: {query}")
        stories = fetch_bing_search(query, count)
        if not stories:
            print("Bing returned empty, trying akshare...")
            stories = fetch_akshare_web(query, count)
        print(f"Found {len(stories)} results")
        return stories
    else:
        print(f"Unknown mode: {mode}")
        return []


def save(stories: list[dict], prefix: str = "next_web"):
    """Save stories to JSON, CSV, and analysis files."""
    json_path = f"{prefix}.json"
    csv_path = f"{prefix}.csv"
    analysis_path = f"{prefix}_analysis.json"

    with open(json_path, "w") as f:
        json.dump(stories, f, indent=2)
    print(f"Saved {len(stories)} stories to {json_path}")

    pd.DataFrame(stories).to_csv(csv_path, index=False)
    print(f"Saved to {csv_path}")

    stats = analyze_feed(stories)
    with open(analysis_path, "w") as f:
        json.dump(stats, f, indent=2)

    print(f"\nAnalysis Summary:")
    print(f"  Total: {stats['total_stories']}")
    print(f"  Avg score: {stats['score_stats']['mean']:.1f}")
    print(f"  Top domains: {list(stats['top_domains'].keys())[:5]}")
    print(f"  Hot keywords: {list(stats['hot_keywords'].keys())[:10]}")
    return stats


def main():
    parser = argparse.ArgumentParser(description="Dual-mode web feed pipeline")
    parser.add_argument(
        "--mode", choices=["general", "targeted"], default="general",
        help="Mode: general (HN top stories) or targeted (keyword search)"
    )
    parser.add_argument("--query", default="", help="Search query for targeted mode")
    parser.add_argument("--count", type=int, default=50, help="Number of results")
    parser.add_argument(
        "--prefix", default="next_web",
        help="Output file prefix (default: next_web)"
    )
    args = parser.parse_args()

    stories = fetch(mode=args.mode, query=args.query, count=args.count)
    if not stories:
        print("No stories fetched.")
        return

    save(stories, prefix=args.prefix)


if __name__ == "__main__":
    main()
