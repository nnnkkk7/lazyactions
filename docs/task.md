# GitHub Actions TUI ツール 要件定義書

## プロジェクト概要

### プロジェクト名（仮）
**lazyactions** - lazygit/k9s スタイルの GitHub Actions 管理 TUI

### コンセプト
> "lazygit for GitHub Actions - Complete CI/CD management without leaving your terminal"

ターミナルから離れることなく、GitHub Actions のワークフロー実行・監視・管理を完結させる TUI ツール。
lazygit/lazydocker のUI/UXを踏襲し、これらのツールに慣れたユーザーが直感的に使える設計を目指す。

---

## 背景

### 現状の課題

#### 1. 既存ツールの限界

**gama (termkit/gama)** - 現在最も機能的な GitHub Actions TUI
- ⭐ 467 stars, Go + BubbleTea, GPL-3.0
- 機能: ワークフロー一覧、履歴表示、トリガー実行、ライブモード（15秒更新）
- **致命的な欠点**:
  - ログストリーミングなし（実行中に何が起きているか見えない）
  - キャンセル・再実行不可
  - アーティファクト管理なし
  - 単一リポジトリのみ対応
  - Self-hosted Runner 状態確認不可

**gh CLI** - GitHub 公式 CLI
- `gh run list/view/watch/rerun/cancel` 等のコマンドあり
- TUI ではなく、インタラクティブな操作体験がない
- 複数リポジトリの横断的な監視は困難

**act (nektos/act)** - ローカル実行ツール
- GitHub Actions をローカルで Docker 実行
- リモートの実行監視とは別用途

#### 2. 開発者のペインポイント

Hacker News や GitHub Issues から収集した声：

- 「PR を出して Actions の結果を待つ → ブラウザで確認」の往復が面倒
- `gh run watch` はあるが、失敗時にすぐログを見られない
- 複数リポジトリのモノレポ構成で、全体の CI 状況を把握しにくい
- Self-hosted Runner がオフラインになっても気づきにくい
- キャッシュが肥大化しても可視化されていない

#### 3. ユースケース例（Bucketeer 開発での想定）

- 複数の SDK リポジトリ（Go, Android, iOS, Web）の CI を一元監視
- PR 作成後、ターミナルでそのままテスト結果を確認
- 失敗したジョブのログを即座に確認し、必要なら再実行
- Self-hosted Runner の状態をリアルタイムで把握
- 肥大化したキャッシュの確認と削除

---

## 競合分析

### 機能比較マトリクス

| 機能 | gama | gh CLI | 新ツール |
|------|------|--------|----------|
| ワークフロー一覧 | ✅ | ✅ | ✅ |
| 実行履歴 | ✅ | ✅ | ✅ |
| ワークフロー実行 | ✅ | ✅ | ✅ |
| **ログストリーミング** | ❌ | △ (watch) | ✅ |
| **キャンセル** | ❌ | ✅ | ✅ |
| **再実行** | ❌ | ✅ | ✅ |
| **アーティファクト管理** | ❌ | ✅ | ✅ |
| **マルチリポジトリ** | ❌ | ❌ | ✅ |
| **Runner 状態監視** | ❌ | ❌ | ✅ |
| **キャッシュ管理** | ❌ | ✅ (ext) | ✅ |
| **3ペイン TUI** | ❌ | N/A | ✅ |
| インタラクティブ操作 | ✅ | ❌ | ✅ |

### gama の詳細分析

**強み:**
- BubbleTea による洗練された TUI
- JSON 形式で 10+ の workflow_dispatch 入力に対応
- Docker 対応で環境構築が簡単
- ライブモード（自動更新）

**弱み:**
- 「実行する」ことに特化しすぎ、「監視・運用」が弱い
- ログが見えないため、失敗時にブラウザを開く必要がある
- 単一リポジトリ前提の設計

---

## 要件定義

### 確定事項（2026-01-18 要件詰め会議）

#### スコープ決定
| 項目 | 決定 | 備考 |
|------|------|------|
| ターゲットユーザー | 個人開発者 | 単一〜少数リポジトリを効率的に管理 |
| MVP対象リポジトリ | 単一リポジトリのみ | カレントディレクトリから自動検出（lazygit方式） |
| 最優先機能 | ログストリーミング | gamaの最大の弱点を解決 |
| GHES対応 | 不要 | GitHub.comのみ対応 |
| 設定ファイル | MVP後回し | gh CLI + 環境変数のみで動作 |
| アーティファクト | MVP後回し | Phase 2で対応 |

#### 認証方式（優先順）
1. `gh auth token` の出力を利用（推奨・設定不要）
2. 環境変数 `GITHUB_TOKEN`
3. 設定ファイル（Phase 2以降）

#### ログストリーミング実装
- **方式**: ポーリング（2-3秒間隔）
- **理由**: GitHub Actions APIはWebSocket未提供、`gh run watch`も内部的にポーリング
- **最適化**: 実行中のジョブのみポーリング（API制限5000 req/hour考慮）

### 機能要件

#### P0: Must Have（MVP）

| # | 機能 | 説明 | GitHub API |
|---|------|------|------------|
| F01 | ワークフロー一覧 | リポジトリのワークフローを表示 | `GET /repos/{owner}/{repo}/actions/workflows` |
| F02 | 実行一覧 | ワークフロー実行の履歴表示 | `GET /repos/{owner}/{repo}/actions/runs` |
| F03 | ジョブ詳細 | 各ジョブのステップ・ステータス表示 | `GET /repos/{owner}/{repo}/actions/runs/{id}/jobs` |
| F04 | **ログストリーミング** | 実行中ジョブのログをリアルタイム表示（**最優先**） | `GET /repos/{owner}/{repo}/actions/jobs/{id}/logs` |
| F05 | **キャンセル** | 実行中のワークフローを停止 | `POST /repos/{owner}/{repo}/actions/runs/{id}/cancel` |
| F06 | **再実行** | 失敗したワークフロー/ジョブを再実行 | `POST /repos/{owner}/{repo}/actions/runs/{id}/rerun` |
| F07 | ワークフロー実行 | workflow_dispatch でトリガー | `POST /repos/{owner}/{repo}/actions/workflows/{id}/dispatches` |
| F08 | **ブランチ/PRフィルター** | 特定ブランチやPRの実行のみ表示 | クエリパラメータ `branch`, `event` |
| F09 | **クリップボードコピー** | ログをクリップボードにコピー | - |
| F10 | **失敗ステップ自動ジャンプ** | エラー時に失敗箇所へ自動移動 | - |

#### P1: Should Have（Phase 2）

| # | 機能 | 説明 | GitHub API |
|---|------|------|------------|
| F11 | **マルチリポジトリ** | 複数リポジトリを一画面で監視 | 複数 API 呼び出し |
| F12 | **アーティファクト管理** | 一覧・ダウンロード | `GET /repos/{owner}/{repo}/actions/artifacts` |
| F13 | **Runner 状態監視** | Self-hosted Runner のステータス表示 | `GET /repos/{owner}/{repo}/actions/runners` |
| F14 | ワークフロー YAML 表示 | エディタなしで内容確認 | `GET /repos/{owner}/{repo}/actions/workflows/{id}` |
| F15 | ジョブ依存グラフ | `needs` の依存関係を可視化 | Jobs API から構築 |
| F16 | 設定ファイル対応 | `~/.config/lazyactions/config.yaml` | - |

#### P2: Nice to Have（Phase 3）

| # | 機能 | 説明 | GitHub API |
|---|------|------|------------|
| F17 | キャッシュ管理 | 一覧・削除・使用量確認 | `GET /repos/{owner}/{repo}/actions/caches` |
| F18 | 使用量メトリクス | 分数・ストレージ使用量 | `GET /repos/{owner}/{repo}/actions/cache/usage` |
| F19 | シークレット一覧 | シークレット名の確認（値は非表示） | `GET /repos/{owner}/{repo}/actions/secrets` |
| F20 | ワークフロー有効/無効化 | workflow の enable/disable | `PUT /repos/{owner}/{repo}/actions/workflows/{id}/enable` |

### 非機能要件

#### パフォーマンス
- 起動時間: 1秒以内
- API レスポンス表示: 500ms 以内
- ログストリーミング更新間隔: 2秒（設定可能）
- メモリ使用量: 100MB 以下

#### ユーザビリティ
- **lazygitスタイル完全準拠**
  - Vim ライクなキーバインド（j/k でナビゲーション）+ 矢印キー併用
  - Tab / Shift-Tab でパネル移動
  - 下部ステータスバーに現在の操作可能アクション表示
  - `?` キーでポップアップヘルプ表示（lazygit風）
  - 破壊的操作（キャンセル等）前に確認ダイアログ
- カラースキーム: 緑=成功、赤=失敗、黄=実行中（lazygit/lazydocker準拠）
- ログ表示: ステップ単位で展開（Enter/Lで全画面切り替え可能）

#### 互換性
- OS: Linux, macOS, Windows
- Go 1.21+
- GitHub.com のみ対応（GHES対応は予定なし）

#### セキュリティ
- トークンはファイルまたは環境変数で管理
- `gh auth token` との連携で既存認証を再利用
- トークンをログに出力しない

---

## UI/UX 設計

### 画面構成（3ペインレイアウト）

```
┌─────────────────────────────────────────────────────────────────────────┐
│  lazyactions                                    bucketeer/bucketeer     │
├─────────────┬───────────────────┬───────────────────────────────────────┤
│ Workflows   │ Runs              │ Logs                                  │
│             │                   │                                       │
│ > ci.yml    │ > #1234 ● 2m      │ Job: build (running)                  │
│   deploy    │   #1233 ✓ 5m      │ ────────────────────────────────────  │
│   release   │   #1232 ✗ 3m      │ ✓ Set up job                    0s    │
│             │   #1231 ✓ 8m      │ ✓ Checkout                      2s    │
│             │                   │ ✓ Setup Go 1.24                15s    │
│             │                   │ ● Run tests              [running]    │
│             │                   │   > go test ./...                     │
│             │                   │   === RUN   TestFeatureFlag           │
│             │                   │   --- PASS: TestFeatureFlag (0.02s)   │
│             │                   │   === RUN   TestEvaluation            │
│             │                   │   ▌                                   │
├─────────────┴───────────────────┴───────────────────────────────────────┤
│ [t]rigger [c]ancel [r]erun [l]ogs [a]rtifacts [R]unners [?]help         │
└─────────────────────────────────────────────────────────────────────────┘
```

### マルチリポジトリビュー

```
┌─────────────────────────────────────────────────────────────────────────┐
│  lazyactions - Dashboard                          12 active runs        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  bucketeer/bucketeer                                                    │
│    ● ci.yml #1234        running   2m    feature/new-sdk                │
│    ✓ release.yml #1230   success   5m    main                           │
│                                                                         │
│  bucketeer/go-server-sdk                                                │
│    ✗ test.yml #456       failed    3m    fix/evaluation-bug             │
│    ● build.yml #457      running   1m    main                           │
│                                                                         │
│  bucketeer/android-client-sdk                                           │
│    ✓ ci.yml #789         success   8m    main                           │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ [Enter] Select  [/] Filter  [R] Refresh  [?] Help                       │
└─────────────────────────────────────────────────────────────────────────┘
```

### Runner 管理画面

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Self-hosted Runners                              Organization: bucket  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  NAME              OS       STATUS      BUSY    LABELS                  │
│  ─────────────────────────────────────────────────────────────────────  │
│  runner-prod-01   linux    ● online    yes     self-hosted, X64, prod   │
│  runner-prod-02   linux    ● online    no      self-hosted, X64, prod   │
│  runner-dev-01    linux    ● online    no      self-hosted, X64, dev    │
│  mac-m1-build     macos    ○ offline   -       self-hosted, ARM64, mac  │
│                                                                         │
│  Total: 4 runners | Online: 3 | Busy: 1                                 │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ [Enter] Details  [d] Delete  [R] Refresh  [?] Help                      │
└─────────────────────────────────────────────────────────────────────────┘
```

### キーバインド設計（lazygitスタイル準拠）

#### ナビゲーション
| キー | アクション | コンテキスト |
|------|------------|--------------|
| `j` / `↓` | 下移動 | 全画面 |
| `k` / `↑` | 上移動 | 全画面 |
| `Tab` | 次のペインへ | メイン画面 |
| `Shift+Tab` | 前のペインへ | メイン画面 |
| `h` / `l` | ペイン移動（代替） | メイン画面 |
| `Enter` | 選択・展開・全画面ログ | 全画面 |
| `Esc` | 戻る/ポップアップ閉じる | 全画面 |
| `q` | 戻る/終了 | 全画面 |

#### アクション
| キー | アクション | コンテキスト |
|------|------------|--------------|
| `t` | ワークフロー実行 | Workflows ペイン |
| `c` | キャンセル（確認あり） | Runs ペイン |
| `r` | 再実行 | Runs ペイン |
| `L` | ログ全画面表示 | Runs/Jobs ペイン |
| `y` | ログをクリップボードにコピー | Logs ペイン |
| `/` | フィルター（ブランチ/PR） | リスト画面 |
| `?` | ヘルプ（ポップアップ） | 全画面 |

#### Phase 2以降
| キー | アクション | コンテキスト |
|------|------------|--------------|
| `a` | アーティファクト | Runs ペイン |
| `R` | Runner 画面 | 全画面 |
| `C` | キャッシュ画面 | 全画面 |

---

## 技術スタック

### 言語・フレームワーク
- **Go 1.21+**
- **BubbleTea** - TUI フレームワーク（gama と同じ、実績あり）
- **Lipgloss** - スタイリング
- **Bubbles** - UI コンポーネント（list, viewport, textinput 等）

### GitHub API クライアント
- **google/go-github** - 公式 Go クライアント
- または REST API 直接呼び出し

### 認証（優先順）
1. **gh CLI 連携（推奨）**: `gh auth token` の出力を利用
   - 設定不要で最も簡単
   - gh CLI インストール済みなら即座に動作
2. 環境変数: `GITHUB_TOKEN`
3. 設定ファイル: `~/.config/lazyactions/config.yaml`（Phase 2以降）

### 設定ファイル形式

```yaml
# ~/.config/lazyactions/config.yaml
github:
  token: ${GITHUB_TOKEN}  # 環境変数参照も可
  # enterprise_url: https://github.example.com  # GHES 対応

repositories:
  - owner: bucketeer
    repo: bucketeer
  - owner: bucketeer
    repo: go-server-sdk
  - owner: bucketeer
    repo: android-client-sdk

settings:
  refresh_interval: 5s      # 自動更新間隔
  log_stream_interval: 2s   # ログストリーミング間隔
  theme: auto               # auto, light, dark

keys:
  quit: ctrl+c
  help: "?"
  refresh: R
```

---

## GitHub API エンドポイント一覧

### ワークフロー関連
```
GET  /repos/{owner}/{repo}/actions/workflows
GET  /repos/{owner}/{repo}/actions/workflows/{workflow_id}
PUT  /repos/{owner}/{repo}/actions/workflows/{workflow_id}/enable
PUT  /repos/{owner}/{repo}/actions/workflows/{workflow_id}/disable
POST /repos/{owner}/{repo}/actions/workflows/{workflow_id}/dispatches
```

### 実行関連
```
GET  /repos/{owner}/{repo}/actions/runs
GET  /repos/{owner}/{repo}/actions/runs/{run_id}
GET  /repos/{owner}/{repo}/actions/runs/{run_id}/jobs
POST /repos/{owner}/{repo}/actions/runs/{run_id}/cancel
POST /repos/{owner}/{repo}/actions/runs/{run_id}/rerun
POST /repos/{owner}/{repo}/actions/runs/{run_id}/rerun-failed-jobs
DELETE /repos/{owner}/{repo}/actions/runs/{run_id}
```

### ログ関連
```
GET /repos/{owner}/{repo}/actions/jobs/{job_id}/logs
GET /repos/{owner}/{repo}/actions/runs/{run_id}/logs
```

### アーティファクト関連
```
GET    /repos/{owner}/{repo}/actions/artifacts
GET    /repos/{owner}/{repo}/actions/artifacts/{artifact_id}
DELETE /repos/{owner}/{repo}/actions/artifacts/{artifact_id}
GET    /repos/{owner}/{repo}/actions/runs/{run_id}/artifacts
```

### Runner 関連
```
GET    /repos/{owner}/{repo}/actions/runners
GET    /repos/{owner}/{repo}/actions/runners/{runner_id}
DELETE /repos/{owner}/{repo}/actions/runners/{runner_id}
GET    /orgs/{org}/actions/runners
```

### キャッシュ関連
```
GET    /repos/{owner}/{repo}/actions/caches
DELETE /repos/{owner}/{repo}/actions/caches/{cache_id}
DELETE /repos/{owner}/{repo}/actions/caches?key={key}
GET    /repos/{owner}/{repo}/actions/cache/usage
```

---

## 開発フェーズ

### Phase 1: MVP
**目標**: ログストリーミングを中心とした単一リポジトリ対応のTUI

#### 1.1 基盤構築
- [ ] プロジェクトセットアップ（Go module, BubbleTea）
- [ ] gh CLI 連携による認証実装
- [ ] GitHub API クライアント実装（go-github）
- [ ] カレントディレクトリからリポジトリ自動検出

#### 1.2 lazygitスタイルUI
- [ ] 3ペインレイアウト（Workflows / Runs / Logs）
- [ ] Tab/Shift-Tab によるパネル移動
- [ ] j/k + 矢印キーによるナビゲーション
- [ ] 下部ステータスバー（利用可能アクション表示）
- [ ] `?` キーでポップアップヘルプ
- [ ] カラースキーム（緑=成功、赤=失敗、黄=実行中）

#### 1.3 コア機能
- [ ] ワークフロー一覧・実行一覧表示
- [ ] **ログストリーミング（最優先）** - ポーリング2-3秒
- [ ] ステップ単位でのログ展開
- [ ] **失敗ステップ自動ジャンプ**
- [ ] ログ全画面表示（Enter/L）
- [ ] **クリップボードコピー（y）**
- [ ] キャンセル（確認ダイアログ付き）
- [ ] 再実行
- [ ] workflow_dispatch 実行
- [ ] **ブランチ/PRフィルター（/）**

#### 1.4 リリース準備
- [ ] バイナリリリース（goreleaser）
- [ ] Homebrew Formula
- [ ] 基本ドキュメント（README）

### Phase 2: 運用機能
- [ ] マルチリポジトリ対応（ダッシュボードビュー）
- [ ] アーティファクト管理（一覧・ダウンロード）
- [ ] Runner 状態監視
- [ ] 設定ファイル対応（`~/.config/lazyactions/config.yaml`）
- [ ] ワークフロー YAML 表示
- [ ] ジョブ依存グラフ

### Phase 3: 拡張機能
- [ ] キャッシュ管理
- [ ] 使用量メトリクス
- [ ] シークレット一覧
- [ ] ワークフロー有効/無効化

---

## 成功指標

### 定量指標
- GitHub Stars: 6ヶ月で 500+（gama 超え目標）
- 月間ダウンロード数: 1,000+
- Issue 解決率: 80% 以上

### 定性指標
- 「ブラウザを開かずに CI/CD 管理が完結する」体験の実現
- lazygit/lazydocker ユーザーからの「操作感が同じで使いやすい」フィードバック
- PR作成→テスト確認→ログ確認→再実行がターミナル内で完結

---

## 参考リンク

### 競合・参考ツール

- [gama (termkit/gama)](https://github.com/termkit/gama) - 既存の GitHub Actions TUI
- [gh CLI run command](https://cli.github.com/manual/gh_run) - GitHub 公式 CLI

### UI/UX 参考（lazygitスタイル準拠）

- [lazygit](https://github.com/jesseduffield/lazygit) - Git TUI（**主要UI参考**）
- [lazydocker](https://github.com/jesseduffield/lazydocker) - Docker TUI（UI参考）
- [k9s](https://github.com/derailed/k9s) - Kubernetes TUI（UI参考）

### 技術

- [BubbleTea](https://github.com/charmbracelet/bubbletea) - TUI フレームワーク
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - スタイリング
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI コンポーネント
- [GitHub Actions REST API](https://docs.github.com/en/rest/actions)
- [google/go-github](https://github.com/google/go-github) - GitHub API クライアント
