# ytpl - CLI YouTube Music Player and Playlist Manager

[English](README.md)

`ytpl` は、YouTubeから音楽をダウンロードしてローカルで管理・再生できるコマンドライン YouTube ミュージックプレイヤーです。プレイリスト機能により、お気に入りの楽曲を整理して再生することができます。

![Image](https://github.com/user-attachments/assets/8d1abd82-91f6-45dd-9ce5-144bc9221395)

(ビジュアライザーは[cli-visualizer](https://github.com/PosixAlchemist/cli-visualizer)

## 主な機能

- YouTubeからの音楽ダウンロード・再生
- ローカル楽曲の管理
- プレイリストの作成・管理
- 楽曲メタデータの編集（タイトル、アーティスト等）
- シャッフル再生
- 再生状況の確認・制御

## インストール

### 必要な環境

- Go 1.16 以上
- MPV プレイヤー
- yt-dlp

### インストール手順

#### ビルド済みバイナリを使用する場合

[リリースページ](https://github.com/256x/ytpl/releases)からお使いのOSとアーキテクチャに対応したバイナリをダウンロードし、実行可能にしてPATHの通ったディレクトリに配置してください。

```bash
# 例：Linux amd64の場合
wget https://github.com/256x/ytpl/releases/download/vX.Y.Z/ytpl_linux_amd64
chmod +x ytpl_linux_amd64
sudo mv ytpl_linux_amd64 /usr/local/bin/ytpl
```

#### Goを使用する場合

Goがインストールされている場合は、以下のコマンドでインストールできます：

```bash
# Go 1.16以降
go install github.com/256x/ytpl@latest
```

`go install` コマンドは `$GOPATH/bin` または `$GOBIN` にバイナリをインストールします。これらのディレクトリが `PATH` に含まれていることを確認してください。

#### ソースからビルドする場合

```bash
# リポジトリをクローン
git clone https://github.com/256x/ytpl.git
cd ytpl

# 依存関係をダウンロード
go mod download

# ビルド
go build -o ytpl

# バイナリをPATHの通ったディレクトリに移動
sudo mv ytpl /usr/local/bin/
```

## 使用方法

### 基本コマンド

```
# ヘルプの表示
ytpl --help

# YouTubeで音楽を検索してダウンロード・再生
ytpl search [クエリ]
# 例：
# ytpl search "アーティスト名 楽曲名"        # アーティストと楽曲名で検索
# ytpl search "楽曲名 カバー"              # カバー楽曲を検索
# ytpl search "アーティスト名 アルバム名"   # アルバムから検索
# ytpl search "https://youtube.com/..."   # YouTube URLから直接再生
# ytpl search "プレイリスト名"             # プレイリストを検索・再生
# ytpl search "ライブ 楽曲名"              # ライブ音源を検索
# ytpl search "カバー 楽曲名"              # カバー動画を検索

# 楽曲メタデータの編集（タイトル、アーティスト等）
ytpl edit [クエリ]
# 例：
# ytpl edit                  # インタラクティブな楽曲選択
# ytpl edit "楽曲名"        # 特定の楽曲を検索して編集

# ローカル保存楽曲の再生
ytpl play [クエリ]
# 例：
# ytpl play                     # インタラクティブな楽曲選択
# ytpl play "アーティスト名"     # アーティスト名で検索
# ytpl play "楽曲名"            # 楽曲名で検索
# ytpl play "アルバム名"        # アルバム名で検索

# プレイリストの管理・再生
ytpl list

# プレイリストの再生
# ytpl list play <プレイリスト名>    # 順番に再生
# ytpl list shuffle <プレイリスト名> # シャッフル再生

# プレイリストの作成・管理
# ytpl list make <プレイリスト名>     # 新しいプレイリストを作成
# ytpl list add <プレイリスト> <楽曲ID> # プレイリストに追加
# ytpl list remove <プレイリスト> <楽曲ID> # プレイリストから削除
# ytpl list delete <プレイリスト>        # プレイリストを削除

# 全ローカル楽曲のシャッフル再生
ytpl shuffle

# 現在の再生状況を表示
ytpl status

# 再生制御
ytpl play [クエリ]  # ローカル保存楽曲の再生
# 例：
# ytpl play
# ytpl play "アーティスト名 楽曲名" # あいまい検索でリスト表示

ytpl pause   # 一時停止
ytpl resume  # 一時停止から再開
ytpl stop    # 停止
ytpl next    # 次の楽曲にスキップ
ytpl prev    # 前の楽曲に戻る
ytpl volume <0-100>  # 音量設定（0-100）

# ローカルストレージから楽曲を削除
ytpl delete [クエリ]

# バージョン情報の表示
ytpl --version または ytpl -v
```

### プレイリスト管理

```
# インタラクティブなプレイリスト操作（サブコマンドが指定されていない場合）
ytpl list

# 新しいプレイリストを作成
ytpl list create MyPlaylist

# 現在再生中の楽曲をプレイリストに追加
# 指定されたプレイリストが存在しない場合は作成されます
ytpl list add MyPlaylist

# 現在再生中の楽曲をプレイリストから削除
ytpl list remove MyPlaylist

# プレイリストを削除
ytpl list del MyPlaylist

# プレイリストの内容を表示
ytpl list show MyPlaylist

# プレイリストを再生（順番通り）
ytpl list play MyPlaylist

# プレイリストをシャッフル再生
ytpl list shuffle MyPlaylist
```

### 楽曲管理

```
# ダウンロード済み楽曲の削除
# 削除された楽曲は全てのプレイリストから自動的に削除されます
ytpl delete

# 楽曲メタデータの編集（タイトル、アーティスト等）
ytpl edit [クエリ]

# 音量調整（0-100）
ytpl volume 80
```

## 設定

設定ファイルは `~/.config/ytpl/config.toml` に保存されます。以下の設定項目が利用可能で、それぞれデフォルト値があります：

```toml
# YouTube音声ファイルの保存ディレクトリ
# $HOME等の環境変数を使用可能
download_dir = "$HOME/.local/share/ytpl/mp3/"

# メディアプレイヤーのパス（mpvを推奨）
player_path = "mpv"

# MPV IPC（プロセス間通信）ソケットパス
player_ipc_socket_path = "/tmp/ytpl-mpv-socket"

# デフォルト音量（0-100）
default_volume = 80

# yt-dlpのパス
yt_dlp_path = "yt-dlp"

# プレイリスト保存ディレクトリ
playlist_dir = "$HOME/.local/share/ytpl/playlists/"

# クッキーを読み込むブラウザ（例："chrome", "firefox", "chromium", "brave", "edge"）
cookie_browser = "chrome"

# ブラウザプロファイル名（通常は不要）
# cookie_profile = ""

# YouTubeからの検索結果の最大取得数
max_search_results = 30
```

### 主要設定項目の説明

- `download_dir`: ダウンロードした楽曲を保存するディレクトリ
- `player_path`: MPVプレイヤーのパス（`mpv` と指定する場合はPATHに含まれている必要があります）
- `player_ipc_socket_path`: MPV制御に使用するIPCソケットパス
- `default_volume`: 起動時のデフォルト音量（0-100）
- `yt_dlp_path`: yt-dlpのパス（デフォルト: "yt-dlp"）
- `playlist_dir`: プレイリストを保存するディレクトリ（デフォルト: "$HOME/.local/share/ytpl/playlists/"）
- `cookie_browser`: クッキーを読み込むブラウザを指定（ログイン必要な動画のダウンロードに必要、デフォルト: "firefox"）
- `max_search_results`: 検索結果の最大表示数

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。詳細は[LICENSE](LICENSE)ファイルをご覧ください。

## 重要な注意事項

- このソフトウェアはYouTubeの利用規約に従って使用してください。
- ダウンロードしたコンテンツは個人使用の範囲内に留めてください。
- 大量の楽曲をダウンロードする際は、YouTubeのレート制限にご注意ください。

