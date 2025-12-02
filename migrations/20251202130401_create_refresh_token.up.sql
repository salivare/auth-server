CREATE TABLE IF NOT EXISTS refresh_token (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     token_hash TEXT NOT NULL UNIQUE,
     user_id INT NOT NULL,
     expires_at DATETIME NOT NULL,
     create_at DATETIME NOT NULL
);
