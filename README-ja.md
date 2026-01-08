<p align="right"><a href="README.md">🇬🇧English</a> | <a href="README-ja.md">🇯🇵日本語</a></p>

<div align="center">

# wips-cli

**開発者のためのクイックメモ・軽量ジャーナリングCLIツール**

[![Release](https://img.shields.io/github/v/release/rynskrmt/wips-cli?style=flat)](https://github.com/rynskrmt/wips-cli/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![macOS](https://img.shields.io/badge/macOS-000000?style=flat&logo=apple&logoColor=white)
![Linux](https://img.shields.io/badge/Linux-FCC624?style=flat&logo=linux&logoColor=black)
![Windows](https://img.shields.io/badge/Windows-0078D6?style=flat&logo=windows&logoColor=white)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)

<p align="center" style="margin: 20px 0;">
    <img src="assets/interactive-mode.gif" alt="動作デモ" width="800">
    <br>
    <em>基本的な使い方</em>
</p>

</div>

## wips-cliとは？

`wips-cli`は、開発メモをサッと記録できるCLIツールです。すべてローカルに保存される、あなただけの開発記録です。

**wips** は**WIP** (Work In Progress)の複数形です。通常のコミット履歴には残らない、開発中の細かな作業状態や思考の流れを記録できます。

## 動作要件

Gitコミットの自動記録機能を使うには、`git`のインストールが必要です。

## インストール

以下のいずれかの方法でインストールできます。

| 方法    | コマンド                                                                                   |
| ------- | ------------------------------------------------------------------------------------------ |
| brew    | `brew install rynskrmt/tap/wips`                                                           |
| scoop   | `scoop bucket add rynskrmt https://github.com/rynskrmt/scoop-bucket && scoop install wips` |
| release | [GitHub Releases](https://github.com/rynskrmt/wips-cli/releases) からダウンロード          |
| curl    | `curl -sfL https://raw.githubusercontent.com/rynskrmt/wips-cli/main/install.sh \| sh`      |
| go      | `go install github.com/rynskrmt/wips-cli/cmd/wip@latest`                                   |
| build   | リポジトリをクローンして `make dev` を実行                                                 |

### 手動インストール (バイナリ)

1. [GitHub Releases](https://github.com/rynskrmt/wips-cli/releases) から、お使いのOSとアーキテクチャに合った最新のバイナリをダウンロードします。
2. アーカイブを解凍します。
3. `wip` バイナリを `PATH` の通ったディレクトリ（例：`/usr/local/bin` や `~/bin`）に移動します。
4. 実行権限を付与します： `chmod +x /path/to/wip`


## 基本的な使い方

メッセージを付けて実行するだけで、現在のコンテキスト（Gitリポジトリ、ブランチ、ディレクトリ）を自動検出してメモを記録します。

```shell
wip "明日は二つ目の機能を実装したい!"
```

## コマンド一覧

メモの管理や作業サマリーの表示に使えるコマンドを用意しています。

```shell
wip [command]
```

各コマンドの説明

| コマンド  | エイリアス | 説明                                                 |
| --------- | ---------- | ---------------------------------------------------- |
| `summary` | `sum`      | 指定期間（日次・週次・カスタム）の作業サマリーを表示 |
| `search`  |            | 自然言語での日付指定や正規表現でイベントを検索       |
| `tail`    | `t`        | 現在のディレクトリでの最近のイベントを表示           |
| `edit`    | `e`        | イベントをIDで編集（デフォルト：最新）               |
| `delete`  |            | イベントをIDで削除（デフォルト：最新）               |
| `hooks`   |            | Gitフック連携の管理（コミットの自動記録）            |
| `sync`    |            | 外部ツール（Obsidian等）へのログ同期                 |
| `config`  |            | グローバル設定の管理                                 |

## 直近の記録を確認

現在のディレクトリでの作業履歴を確認

```shell
$ wip tail   # または: wip t
```

<p align="center" style="margin: 20px 0;">
    <img src="assets/tail.gif" alt="直近の記録を確認できます" width="800">
    <br>
    <em>直近の記録を確認できます</em>
</p>

`-g`で全プロジェクトの履歴を表示、`-n`で件数を指定できます。

```shell
$ wip tail -g      # 全プロジェクトの履歴
$ wip tail -n 20   # 直近20件を表示
```

## 作業サマリー

今日の作業内容を確認するには

```shell
$ wip summary   # または: wip sum
```

<p align="center" style="margin: 20px 0;">
    <img src="assets/summary.gif" alt="サマリーのデモ" width="800">
    <br>
    <em>1日のメモをサッと確認できます</em>
</p>

過去の日付や週単位でも確認できます

```shell
$ wip sum --week       # 今週
$ wip sum --last-week  # 先週
$ wip sum --days 3     # 過去3日分
```

### エクスポート

サマリーを各種形式でファイル出力できます

```shell
$ wip sum --week --format md --out report.md
```

## Obsidian同期機能 (Experimental)

外部ツール（Obsidian等）と日々のログを同期できます。

### Obsidianの設定

まず、Obsidianのデイリーノート（Daily Note）フォルダのパスを設定します：

```shell
$ wip config sync obsidian enable --path "~/ObsidianVault/Daily"
```

### 同期の実行

今日のログ（ディレクトリごとにグループ化）をデイリーノートに同期するには、以下を実行します：

```shell
$ wip sync
```

デイリーノートの特定セクション（デフォルト：`## wips-cli logs`）にログを追記します。
何度実行しても内容は重複せず、該当セクションが最新の状態に更新されます。

### オプション

- `--days <N>`: 過去N日分のログを同期します。まとめて同期したい場合に便利です。
  ```shell
  $ wip sync --days 3
  ```
- `--all`: 過去のすべての履歴（現在は直近365日分）を同期します。
  ```shell
  $ wip sync --all
  ```
- `--create`: デイリーノートが存在しない場合に新規作成します。（デフォルト：存在しない日付はスキップ）
  ```shell
  $ wip sync --create
  ```

## 検索機能

自然言語での日付指定や、フィルタを使った検索が可能です。

```shell
$ wip search "auth bug" --from "last week"
```


## Git連携

リポジトリ内で以下を実行すると、コミットが自動記録されるようになります

```shell
$ wip hooks install
```

インストール後は、`git commit`するたびに自動的に`wips-cli`に記録されます。

## 設定

特定のディレクトリ（例：秘密のプロジェクト）をサマリーから除外するには

```shell
$ wip config add-hidden /path/to/secret-project
```

現在の設定は `wip config list` で確認できます。

## ライセンス

MIT © [rynskrmt](https://github.com/rynskrmt)
