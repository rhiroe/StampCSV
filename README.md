# StampCSV

ボタン1つでタイムスタンプをCSVに記録する勤怠管理アプリです。  
GUIボタンまたはSlack Botのメンションで「開始」「終了」を記録し、月ごとのCSVファイルに自動集計します。

## 機能

- **GUIボタン** によるワンクリック打刻（IN / OUT）
- **保存先ディレクトリ** のフォルダダイアログ選択
- **月ごとのCSVファイル** を自動生成（`YYYY-MM.csv`）
- 当日の合計時間・月間時間計・出勤日数を自動計算
- 日付をまたぐ場合（深夜0時）は前日の終了に `23:59`、当日の開始に `0:00` を自動記録
- **Slack Bot連携**（Socket Mode）— `@StampCSV in` / `@StampCSV out` で打刻

## 動作環境

- Windows / macOS / Linux
- Go 1.21 以上

## インストール

```bash
git clone https://github.com/yourname/StampCSV.git
cd StampCSV
go mod tidy
```

## ビルド

```bash
go build -o StampCSV .
```

> **Windows の場合:** GUIアプリとして起動するには `-ldflags "-H windowsgui"` を付けてビルドすると
> コンソールウィンドウが表示されません。
> ```bash
> go build -ldflags "-H windowsgui" -o StampCSV.exe .
> ```

## 使い方

### GUIアプリ

```bash
./StampCSV
```

1. **「選択」ボタン** を押してCSVを保存するフォルダを選ぶ
2. **「IN（開始）」** ボタンを押すと現在時刻を開始として記録
3. **「OUT（終了）」** ボタンを押すと現在時刻を終了として記録

### Slack Bot

Slack App を作成し、Socket Mode を有効にしたうえで以下の環境変数を設定します。

| 環境変数 | 値 |
|---|---|
| `SLACK_APP_TOKEN` | `xapp-` で始まるアプリレベルトークン |
| `SLACK_BOT_TOKEN` | `xoxb-` で始まるBotトークン |

```bash
export SLACK_APP_TOKEN=xapp-xxxxxxxxxxxxx
export SLACK_BOT_TOKEN=xoxb-xxxxxxxxxxxxx
./StampCSV
```

Slackチャンネルで以下のようにメンションすると打刻できます。

```
@StampCSV in    # 開始を記録
@StampCSV out   # 終了を記録
```

> **Slack Appの設定で必要なスコープ:**  
> Bot Token Scopes: `app_mentions:read`, `chat:write`  
> Event Subscriptions: `app_mention`（Socket Modeでは公開URLは不要）

## CSVフォーマット

ファイル名: `YYYY-MM.csv`（例: `2026-04.csv`）

```
日,時間,開始,終了,開始,終了,開始,終了,開始,終了,開始,終了,休憩等,
1,3:57,20:02,23:59,,,,,,,,,,
2,2:06,0:00,1:34,23:27,23:59,,,,,,,,
...
時間計,93:44,,,,,,,,,,,,
日数,25,,,,,,,,,,,,
時間修正枠,,
```

- 時刻は `H:MM` 形式（例: `9:05`）
- 1日あたり最大5区間（開始/終了ペア）を記録可能
- 「時間」列・「時間計」・「日数」行はスタンプのたびに自動更新

## ディレクトリ構成

```
StampCSV/
├── main.go          # エントリーポイント
├── ui/
│   └── window.go    # Fyneウィンドウ定義
├── csv/
│   └── writer.go    # CSV読み書き・計算ロジック
├── slack/
│   └── bot.go       # Slack Socket Mode Bot
├── go.mod
└── go.sum
```

## 依存ライブラリ

| ライブラリ | 用途 |
|---|---|
| [fyne.io/fyne/v2](https://fyne.io/) | クロスプラットフォームGUI |
| [github.com/slack-go/slack](https://github.com/slack-go/slack) | Slack Bot（Socket Mode） |
