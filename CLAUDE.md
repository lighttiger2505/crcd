# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 概要

`crcd`(Chrome Chrono Drive) は Google Chrome の閲覧履歴・ブックマークを一覧して fzf で選択し、選んだ URL を OS 標準ブラウザで開く CLI ツール。

## コマンド

- `make build` — `go build`(ldflags で version/revision/goversion を埋め込み)
- `make install` — `go install`
- `make test` — `go test -v`
- 単一テスト — `go test -run <TestName> -v`（注: 現状 `_test.go` は未整備）
- `make cover` — カバレッジ取得 + HTML 表示
- `make lint` — `gometalinter`（`lint-config.json` と外部ツールの導入が前提）

注意: `github.com/mattn/go-sqlite3` を使うため **CGO 必須**(C コンパイラが必要)。

## アーキテクチャ

CLI 基盤は `urfave/cli` v1。エントリは `main.go` の `newApp()`。サブコマンドは3つ:

- `history`(alias `s`) → `history()` / `history.go`、`--range,-r` フラグ（例 `1y2m3d` の相対期間）
- `bookmark`(alias `b`) → `bookmark()` / `bookmark.go`
- `config`(alias `c`) → `config()` / `config.go`（`$EDITOR` で設定ファイルを開く）

### 共通フロー（history / bookmark）

エントリ収集 → 表示用の色付き行を生成（1行目=タイトル/名前、2行目=URL）→ `fzfOpen()`(`fzf.go`) で対話選択 → `openbrowser()`(`browser.go`) で選択 URL を OS 標準ブラウザで開く。

### 設定（config.go）

**TOML** 形式、`~/.config/crcd/config.toml`(Windows は `%APPDATA%/crcd`)。フィールドは `Profile` のみ（既定 `"Default"`）。ファイルが無ければ既定値で自動生成される。

### Chrome パス解決（chrome.go）

`getChromeProfilePath(GOOS)` が OS 別のベースディレクトリ + `cfg.Profile` を返し、`getHistoryPath` / `getBookmarkPath` が `History` / `Bookmarks` を付与する。

### 履歴取得（history.go + sqlite.go）

Chrome は稼働中の History DB をロックするため、`copyHisotryDB` で `/tmp` にコピーしてから `selectHistory` で `urls` テーブルを読む。タイムスタンプは WebKit 形式(1601年起点のμs)で `webkitToTime` により変換。並び順は **Frecency**(= `VisitCount` × recency weight。`getRecencyWeight` が経過時間で重み付け)。Frecency 順を維持するため fzf は `--no-sort` で起動する。

### ブックマーク取得（bookmark.go）

Chrome の Bookmarks JSON を読み、`bookmark_bar` を再帰走査して URL エントリ（フォルダパス付き）を収集する。

### fzf（fzf.go）

外部バイナリではなく **`junegunn/fzf` を Go ライブラリとして** 使用。input/output チャネル経由で接続し、選択行の2行目から URL を抽出する。

### 補足

`align.go`(`Format` / `padFields` 等のテーブル整形ユーティリティ群)は現在どのコマンドパスからも呼ばれておらず、未使用の可能性が高い。
