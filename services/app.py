"""
A-share full market data service using akshare.
Provides all A-shares (5000+) for the frontend to use.

Note: This service requires internet connection to fetch data from akshare API.
If the API is unavailable, it will return an error message.
"""

from flask import Flask, jsonify
import akshare as ak

app = Flask(__name__)


@app.route('/health')
def health():
    return jsonify({'status': 'ok', 'message': 'A-share data service is running'})


@app.route('/api/stocks/all')
def get_all_stocks():
    """Get all A-share stocks from akshare."""
    try:
        # Get all A-shares with basic info (5500+ stocks)
        stock_df = ak.stock_info_a_code_name()

        # Format data for frontend
        stocks = []
        for _, row in stock_df.iterrows():
            stocks.append({
                'code': row['code'],
                'name': row['name'],
            })

        return jsonify({
            'stocks': stocks,
            'total': len(stocks)
        })

    except Exception as e:
        # Return error message with helpful context
        return jsonify({
            'error': str(e),
            'message': (
                'Unable to fetch full market data from akshare. '
                'This may be due to a temporary API issue or network connection. '
                'The frontend will use the hardcoded stock list (77 stocks) as fallback.'
            )
        }), 503


@app.route('/api/stocks/search')
def search_stocks():
    """Search stocks by name or code."""
    query = request.args.get('q', '')

    if not query:
        return jsonify({'stocks': [], 'total': 0})

    try:
        # Get all A-shares with basic info (5500+ stocks)
        stock_df = ak.stock_info_a_code_name()

        # Filter by name or code
        filtered = stock_df[
            stock_df['name'].str.contains(query) |
            stock_df['code'].str.contains(query)
        ]

        stocks = [
            {
                'code': row['code'],
                'name': row['name'],
            }
            for _, row in filtered.iterrows()
        ]

        return jsonify({
            'stocks': stocks[:20],  # Return top 20 results
            'total': len(stocks)
        })

    except Exception as e:
        return jsonify({
            'error': str(e),
            'message': 'Unable to search stocks. The frontend will use the hardcoded stock list.'
        }), 503


if __name__ == '__main__':
    app.run(port=8081, debug=True)
