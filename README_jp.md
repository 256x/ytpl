# ytpl - コマンドラインYouTube音楽プレーヤー

[English version](README.md)


`ytpl`は、コマンドラインで動作するYouTube音楽プレーヤーです。YouTubeから音楽をダウンロードし、ローカルで管理・再生することができます。プレイリスト機能を備えており、お気に入りの曲を整理して再生できます。

## 主な機能

- YouTubeからの音楽ダウンロードと再生
- ローカルトラックの管理
- プレイリストの作成・管理
- シャッフル再生
- 再生状態の確認と制御

## インストール

### 前提条件

- Go 1.16 以上
- MPVプレーヤー
- yt-dlp

### インストール手順

#### リリース済みバイナリを使用する場合

[リリースページ](https://github.com/256x/ytpl/releases)からお使いのOSとアーキテクチャに合ったバイナリをダウンロードし、実行可能な状態にしてPATHの通ったディレクトリに配置してください。

```bash
# 例: Linux amd64 の場合
wget https://github.com/256x/ytpl/releases/download/vX.Y.Z/ytpl_linux_amd64
chmod +x ytpl_linux_amd64
sudo mv ytpl_linux_amd64 /usr/local/bin/ytpl
```

#### Go を使用する場合

Go がインストールされている環境では、以下のコマンドでインストールできます：

```bash
# Go 1.16 以降の場合
go install github.com/256x/ytpl@latest
```

`go install` コマンドは、バイナリを `$GOPATH/bin` または `$GOBIN` にインストールします。これらのディレクトリが `PATH` に含まれていることを確認してください。

#### ソースからビルドする場合

```bash
# リポジトリをクローン
git clone https://github.com/256x/ytpl.git
cd ytpl

# 依存関係を取得
go mod download

# ビルド
go build -o ytpl

# バイナリをPATHの通ったディレクトリに配置
sudo mv ytpl /usr/local/bin/
```

## 使い方

### 基本コマンド

```
# ヘルプを表示
ytpl --help

# YouTubeで音楽を検索してダウンロード・再生
ytpl search [query]
# 例: ytpl search "アーティスト名 曲名"      # アーティストと曲名で検索
# 例: ytpl search "曲名 カバー"            # カバー曲を検索
# 例: ytpl search "アーティスト名 アルバム名" # アルバムから検索
# 例: ytpl search "https://youtube.com/..." # YouTubeのURLから直接再生
# 例: ytpl search "プレイリスト名"          # プレイリストを検索して再生
# 例: ytpl search "ライブ 曲名"            # ライブ音源を検索
# 例: ytpl search "歌ってみた 曲名"         # カバー動画を検索

# ローカルに保存されている楽曲を再生
ytpl play [query]
# 例: ytpl play                     # 対話的に楽曲を選択
# 例: ytpl play "アーティスト名"     # アーティスト名で検索
# 例: ytpl play "曲名"              # 曲名で検索
# 例: ytpl play "アルバム名"         # アルバム名で検索

# プレイリストを管理・再生
ytpl list

# プレイリストを再生
# ytpl list play <playlist_name>    # 順番に再生
# ytpl list shuffle <playlist_name> # シャッフル再生

# プレイリストの作成と管理
# ytpl list make <playlist_name>     # 新規作成
# ytpl list add <playlist> <track_id> # 追加
# ytpl list remove <playlist> <track_id> # 削除
# ytpl list delete <playlist>        # 削除


# ローカルの全曲をシャッフル再生
ytpl shuffle

# 現在の再生状態を表示
ytpl status

# 再生コントロール
ytpl play [query]  # ローカルに保存されている楽曲を再生
# 例: ytpl play
# 例: ytpl play "アーティスト名 曲名" #あいまい検索でリスト表示される

ytpl pause   # 再生を一時停止
ytpl resume  # 一時停止から再生を再開
ytpl stop    # 再生を停止
ytpl next    # 次の曲にスキップ
ytpl prev    # 前の曲に戻る
ytpl volume <0-100>  # 音量を設定 (0-100)

# ローカルから楽曲を削除
ytpl delete [query]

# バージョン情報を表示
ytpl --version または ytpl -v
```

### プレイリスト管理

```
# 対話型でプレイリストを操作（サブコマンドを指定しない場合）
ytpl list

# 新しいプレイリストを作成
ytpl list make マイプレイリスト

# 現在再生中の曲をプレイリストに追加
# 指定したプレイリストが存在しない場合は新規作成されます
ytpl list add マイプレイリスト

# 現在再生中の曲をプレイリストから削除
ytpl list remove マイプレイリスト

# プレイリストを削除
ytpl list del マイプレイリスト


# プレイリストを再生
ytpl list play マイプレイリスト

# プレイリストをシャッフル再生
ytpl list shuffle マイプレイリスト
```

### トラック管理

```
# ダウンロードしたトラックを削除
# 削除したトラックはすべてのプレイリストからも自動的に削除されます
ytpl delete

# ボリュームを調整 (0-100)
ytpl volume 80
```

## 設定

設定ファイルは `~/.config/ytpl/config.toml` に保存されます。以下の設定項目が利用可能で、それぞれデフォルト値が設定されています：

```toml
# YouTubeの音声ファイルを保存するディレクトリ
# $HOME などの環境変数が使用可能
download_dir = "$HOME/.local/share/ytpl/mp3/"

# メディアプレーヤーのパス（mpvが推奨）
player_path = "mpv"

# MPVのIPC（プロセス間通信）ソケットのパス
player_ipc_socket_path = "/tmp/ytpl-mpv-socket"

# デフォルトのボリューム（0-100）
default_volume = 80

# yt-dlpのパス
yt_dlp_path = "yt-dlp"

# プレイリストを保存するディレクトリ
playlist_dir = "$HOME/.local/share/ytpl/playlists/"

# クッキーを読み込むブラウザ（例: "chrome", "firefox", "chromium", "brave", "edge"）
cookie_browser = "chrome"

# ブラウザのプロファイル名（通常は不要）
# cookie_profile = ""

# YouTubeから取得する検索結果の最大数
max_search_results = 30
```

### 主な設定項目の説明

- `download_dir`: ダウンロードしたトラックの保存先ディレクトリ
- `player_path`: MPVプレーヤーのパス（`mpv` と指定するとパスが通っている必要があります）
- `player_ipc_socket_path`: MPVの制御に使用するIPCソケットのパス
- `default_volume`: 起動時のデフォルトボリューム（0-100）
- `yt_dlp_path`: yt-dlpのパス（デフォルト: "yt-dlp"）
- `playlist_dir`: プレイリストを保存するディレクトリ（デフォルト: "$HOME/.local/share/ytpl/playlists/"）
- `cookie_browser`: クッキーを読み込むブラウザを指定（ログインが必要な動画のダウンロードに必要、デフォルト: "firefox"）
- `max_search_results`: 検索結果の最大表示件数

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。詳細は[LICENSE](LICENSE)ファイルを参照してください。

## 注意事項

- このソフトウェアは、YouTubeの利用規約に準拠して使用してください。
- ダウンロードしたコンテンツは、個人的な使用目的に限定してください。
- 大量のトラックをダウンロードする場合は、YouTubeのレート制限に注意してください。
