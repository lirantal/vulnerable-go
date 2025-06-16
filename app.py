from flask import Flask, request, render_template_string
import sqlite3
import os

app = Flask(__name__)

# Simple search feature
@app.route('/search', methods=['GET'])
def search():
    query = request.args.get('q', '')

    with sqlite3.connect("database.db") as conn:
        conn.row_factory = sqlite3.Row
        results = conn.execute(f"SELECT * FROM posts WHERE title LIKE '{query}'").fetchall()

    return render_template_string(f"""
        <h2>Results for query "query":</h2>
        <p>{results}</p>
    """)

if __name__ == '__main__':
    app.run(debug=True)