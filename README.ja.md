<div align="center">
  <h1>:musical_note: ytpl</h1>
  <p>
    <a href="README.md">English</a> | <a href="README.ja.md">日本語</a>
  </p>
  <p>プレイリスト管理機能付きコマンドラインYouTubeミュージックプレイヤー</p>
  <p>
    <a href="#特徴">特徴</a> •
    <a href="#インストール">インストール</a> •
    <a href="#使い方">使い方</a> •
    <a href="#設定">設定</a> •
    <a href="#ライセンス">ライセンス</a>
  </p>
</div>

## ✨ 特徴

- ターミナルから直接YouTubeの音楽を検索・再生
- 音楽のダウンロードとローカル管理
- プレイリストの作成と管理
- シンプルで直感的なインターフェース
- 軽量で高速な動作

## 🚀 インストール

### 前提条件

- Go 1.24 以上
- yt-dlp
- mpv プレイヤー

### go install を使用する場合

```bash
go install github.com/256x/ytpl@latest
```

### 手動ビルド

```bash
git clone https://github.com/256x/ytpl.git
cd ytpl
go build -o ytpl .
sudo mv ytpl /usr/local/bin/
```

## 🎮 使い方

### ローカル再生

`play` コマンドは、すでにローカルにダウンロードされた楽曲を再生するために使用します。あいまい検索をサポートしており、部分一致で楽曲を検索できます。

### 検索と再生

```bash
# 曲を検索する (YouTubeで使える検索語句がすべて利用可能)
ytpl search "アーティスト名 曲名"

# ローカルにダウンロード済みの曲を再生 (あいまい検索可能)
ytpl play "アーティスト名 曲名"

# 正確なファイル名でローカル曲を再生
ytpl play "正確なファイル名"
```

### シャッフル再生

```bash
# ローカルにあるすべての曲をシャッフル再生
ytpl shuffle

# 検索に一致する曲をシャッフル再生
ytpl shuffle "検索語句"
```

### プレイリスト管理

```bash
# 新しいプレイリストを作成
ytpl list create プレイリスト名

# プレイリストに曲を追加
ytpl list add プレイリスト名 動画ID

# プレイリスト一覧を表示
ytpl list

# プレイリストを再生
ytpl list play プレイリスト名
```

### プレイヤー操作

```bash
# 再生/一時停止
ytpl pause

# 再生再開
ytpl resume

# 停止
ytpl stop

# 次の曲
ytpl next

# 前の曲
ytpl prev

# 音量調節 (0-100)
ytpl vol 80
```

## ⚙️ 設定

設定ファイルは `~/.config/ytpl/config.toml` に配置されます。

設定例:

```toml
# ダウンロードした音楽ファイルの保存先
download_dir = "~/.local/share/ytpl/mp3/"

# メディアプレイヤー (mpv) のパス
player_path = "mpv"

# デフォルトの音量 (0-100)
default_volume = 80

# yt-dlp のパス
yt_dlp_path = "yt-dlp"

# プレイリストファイルの保存先
playlist_dir = "~/.local/share/ytpl/playlists/"

# クッキーに使用するブラウザ (オプション)
# cookie_browser = "firefox"

# 検索結果の最大表示件数
max_search_results = 15
```

## 📜 ライセンス

このプロジェクトは MIT ライセンスの下で公開されています。詳細は [LICENSE](LICENSE) ファイルを参照してください。

---

<div align="center">
  <p>❤️ 作: <a href="https://github.com/256x">256x</a></p>
</div>
