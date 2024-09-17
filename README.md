DBはpostgreを使用

mac環境にて、
ターミナルを起動
dockerで以下のコマンドを実行して、PostgreSQLコンテナを実行する コマンド
docker run --name my-postgres-container -e POSTGRES_USER=suser -e POSTGRES_PASSWORD=spass -e POSTGRES_DB=company -p 5432:5432 -d postgres

1で構築したPostgreの環境に以下のコマンドを実行じてコンテナに入る
docker exec -it my-postgres-container psql -U suser -d company

3.　テーブルを作成

CREATE TABLE employee (
    id SERIAL PRIMARY KEY,  -- 従業員IDの自動生成（主キー）
    name VARCHAR(100) NOT NULL,  -- 名前
    gender VARCHAR(10),  -- 性別（任意の値、またはM/Fなど）
    hire_year INT,  -- 入社年
    address VARCHAR(255),  -- 住所
    department VARCHAR(100),  -- 部署
    others TEXT,  -- その他の情報
    image BYTEA,  -- 画像データ（バイナリデータとして格納）
    email VARCHAR(255) UNIQUE NOT NULL,  -- メールアドレス（ユニークかつ必須）
    password VARCHAR(255) NOT NULL  -- パスワード（必須）
);
