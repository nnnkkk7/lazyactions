# lazyactions 実装設計書

## 設計方針

### 基本原則

1. **lazygitスタイル最優先** - UIの操作感はlazygitに準拠
2. **ログストリーミング中心** - MVPの最重要機能
3. **シンプルな構成** - 過度な抽象化を避け、必要十分な設計
4. **テスト駆動開発** - 機能実装前にテストを書く（カバレッジ70%以上目標）
5. **セキュリティ重視** - トークン・ログのセキュアな取り扱い

### 差別化ポイント（vs gama）

| 機能 | gama | lazyactions |
|------|------|-------------|
| ログ表示 | なし | **リアルタイムストリーミング** |
| Jobs詳細 | なし | **ステップ詳細表示** |
| UI構成 | タブベース | **lazygit風3ペイン** |
| 失敗時UX | ブラウザへ遷移 | **自動ジャンプ＋コピー** |

---

## ディレクトリ構成

```
lazyactions/
├── cmd/
│   └── lazyactions/
│       └── main.go              # エントリポイント
├── app/
│   ├── app.go                   # メインアプリ（tea.Model）
│   ├── app_test.go
│   ├── keymap.go                # キーバインド定義
│   ├── styles.go                # Lipglossスタイル定義
│   ├── messages.go              # BubbleTeaメッセージ型定義
│   ├── commands.go              # tea.Cmd生成（不変値キャプチャ）
│   ├── list.go                  # FilteredList[T]（汎用リスト）
│   ├── ticker.go                # TickerTask（ポーリング）
│   └── logview.go               # LogViewport（Autoscroll付き）
├── github/
│   ├── interface.go             # Client インターフェース（★本番コード）
│   ├── client.go                # GitHubクライアント実装
│   ├── client_test.go
│   ├── types.go                 # 型定義
│   ├── errors.go                # カスタムエラー型
│   ├── sanitize.go              # ログサニタイズ
│   └── mock.go                  # テスト用モック（本番コードに配置）
├── auth/
│   ├── token.go                 # SecureToken、トークン取得
│   └── token_test.go
├── repo/
│   ├── detect.go                # カレントディレクトリからリポジトリ検出
│   ├── detect_test.go
│   └── validate.go              # 入力バリデーション
├── Makefile
├── go.mod
└── go.sum
```

**シンプルな4パッケージ構成**:
- `app`: TUI全体（BubbleTea）
- `github`: GitHub API呼び出し
- `auth`: 認証（gh CLI連携）
- `repo`: Gitリポジトリ検出

---

## 認証（auth/）

task.md確定事項に基づく優先順位:

```go
//auth/token.go

// GetToken はトークンを優先順位に従って取得
// 1. gh auth token（推奨・設定不要）
// 2. 環境変数 GITHUB_TOKEN
// 3. 設定ファイル（Phase 2以降）
func GetToken() (string, error) {
    // 1. gh CLI から取得（最優先）
    if token, err := getFromGhCLI(); err == nil && token != "" {
        return token, nil
    }

    // 2. 環境変数から取得
    if token := os.Getenv("GITHUB_TOKEN"); token != "" {
        return token, nil
    }

    return "", errors.New("no authentication token found: run 'gh auth login' or set GITHUB_TOKEN")
}

func getFromGhCLI() (string, error) {
    cmd := exec.Command("gh", "auth", "token")
    out, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(out)), nil
}
```

---

## セキュリティ設計

### SecureToken（トークン漏洩防止）

```go
//auth/token.go

// SecureToken はトークンを安全にラップする型
// fmt.Stringer を実装し、ログ出力時に値を隠蔽
type SecureToken struct {
    value string
}

// String は "[REDACTED]" を返す（ログ漏洩防止）
func (t SecureToken) String() string {
    return "[REDACTED]"
}

// GoString は "%#v" 出力時も "[REDACTED]" を返す
func (t SecureToken) GoString() string {
    return "[REDACTED]"
}

// Value はトークン値を返す（内部使用のみ）
func (t SecureToken) Value() string {
    return t.value
}

// NewSecureToken はトークンを検証してSecureTokenを作成
func NewSecureToken(token string) (SecureToken, error) {
    if token == "" {
        return SecureToken{}, errors.New("token is empty")
    }
    if !isValidGitHubTokenFormat(token) {
        return SecureToken{}, errors.New("invalid token format")
    }
    return SecureToken{value: token}, nil
}

func isValidGitHubTokenFormat(token string) bool {
    // ghp_ (classic PAT), github_pat_ (fine-grained), gho_ (OAuth), ghs_ (server-to-server)
    validPrefixes := []string{"ghp_", "github_pat_", "gho_", "ghs_"}
    for _, prefix := range validPrefixes {
        if strings.HasPrefix(token, prefix) {
            return true
        }
    }
    // 従来形式（40文字の16進数）も許可
    return len(token) == 40
}
```

### ログサニタイズ（機密情報除去）

ワークフローログにはシークレットが含まれる可能性があるため、表示前にサニタイズする。

```go
//github/sanitize.go

var secretPatterns = []*regexp.Regexp{
    // GitHub tokens
    regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
    regexp.MustCompile(`github_pat_[a-zA-Z0-9]{22}_[a-zA-Z0-9]{59}`),
    regexp.MustCompile(`gho_[a-zA-Z0-9]{36}`),
    // AWS
    regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
    regexp.MustCompile(`(?i)(aws_secret_access_key|aws_access_key_id)\s*[=:]\s*['"]?[A-Za-z0-9/+=]{20,}['"]?`),
    // Generic secrets
    regexp.MustCompile(`(?i)(api[_-]?key|apikey|secret|password|token|credential|auth)[=:]\s*['"]?[^\s'"]{8,}['"]?`),
}

// SanitizeLogs はログから機密情報を除去
func SanitizeLogs(logs string) string {
    result := logs
    for _, pattern := range secretPatterns {
        result = pattern.ReplaceAllString(result, "[REDACTED]")
    }
    return result
}

// ContainsPotentialSecrets はログに機密情報が含まれる可能性を判定
func ContainsPotentialSecrets(content string) bool {
    for _, pattern := range secretPatterns {
        if pattern.MatchString(content) {
            return true
        }
    }
    return false
}
```

### 入力バリデーション

```go
//repo/validate.go

var (
    validOwnerName    = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$`)
    validRepoName     = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,100}$`)
    validWorkflowPath = regexp.MustCompile(`^\.github/workflows/[a-zA-Z0-9._-]+\.(yml|yaml)$`)
)

func ValidateOwner(owner string) error {
    if !validOwnerName.MatchString(owner) {
        return fmt.Errorf("invalid owner name: %q", owner)
    }
    return nil
}

func ValidateRepoName(name string) error {
    if !validRepoName.MatchString(name) {
        return fmt.Errorf("invalid repository name: %q", name)
    }
    return nil
}
```

---

## リポジトリ検出（repo/）

lazygit方式: カレントディレクトリから自動検出

```go
//repo/detect.go

// Detect はカレントディレクトリからGitHubリポジトリを検出
func Detect() (*github.Repository, error) {
    // 1. .git ディレクトリの存在確認
    if _, err := os.Stat(".git"); os.IsNotExist(err) {
        return nil, errors.New("not a git repository")
    }

    // 2. remote origin の URL を取得
    cmd := exec.Command("git", "remote", "get-url", "origin")
    out, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get remote URL: %w", err)
    }

    // 3. URL をパース（SSH/HTTPS両対応）
    return parseGitHubURL(strings.TrimSpace(string(out)))
}

// parseGitHubURL はGitHub URLからowner/repoを抽出
// 対応形式:
//   - https://github.com/owner/repo.git
//   - git@github.com:owner/repo.git
func parseGitHubURL(url string) (*github.Repository, error) {
    // SSH形式: git@github.com:owner/repo.git
    if strings.HasPrefix(url, "git@github.com:") {
        path := strings.TrimPrefix(url, "git@github.com:")
        path = strings.TrimSuffix(path, ".git")
        parts := strings.Split(path, "/")
        if len(parts) != 2 {
            return nil, fmt.Errorf("invalid GitHub SSH URL: %s", url)
        }
        return &github.Repository{Owner: parts[0], Name: parts[1]}, nil
    }

    // HTTPS形式: https://github.com/owner/repo.git
    if strings.Contains(url, "github.com") {
        u, err := neturl.Parse(url)
        if err != nil {
            return nil, err
        }
        path := strings.TrimPrefix(u.Path, "/")
        path = strings.TrimSuffix(path, ".git")
        parts := strings.Split(path, "/")
        if len(parts) != 2 {
            return nil, fmt.Errorf("invalid GitHub HTTPS URL: %s", url)
        }
        return &github.Repository{Owner: parts[0], Name: parts[1]}, nil
    }

    return nil, fmt.Errorf("not a GitHub repository: %s", url)
}
```

---

## コア型定義

###github/types.go

```go
package github

import "time"

// Repository はリポジトリ情報
type Repository struct {
    Owner string
    Name  string
}

func (r Repository) FullName() string {
    return r.Owner + "/" + r.Name
}

// Workflow はワークフロー定義
type Workflow struct {
    ID    int64
    Name  string
    Path  string  // .github/workflows/ci.yml
    State string  // active, disabled
}

// Run はワークフロー実行
type Run struct {
    ID         int64
    Name       string
    Status     string    // queued, in_progress, completed
    Conclusion string    // success, failure, cancelled
    Branch     string
    Event      string    // push, pull_request, workflow_dispatch
    CreatedAt  time.Time
    Actor      string
    URL        string
}

func (r Run) IsRunning() bool {
    return r.Status == "in_progress" || r.Status == "queued"
}

func (r Run) IsFailed() bool {
    return r.Conclusion == "failure"
}

// Job はジョブ情報
type Job struct {
    ID         int64
    Name       string
    Status     string
    Conclusion string
    Steps      []Step
}

// Step はステップ情報
type Step struct {
    Name       string
    Status     string
    Conclusion string
    Number     int
}

func (s Step) IsFailed() bool {
    return s.Conclusion == "failure"
}
```

###github/errors.go（カスタムエラー型）

```go
package github

// ErrorType はエラーの種類を表す
type ErrorType int

const (
    ErrTypeNetwork ErrorType = iota
    ErrTypeAuth
    ErrTypeRateLimit
    ErrTypeNotFound
    ErrTypeServer
    ErrTypeUnknown
)

// AppError はアプリケーションエラー
type AppError struct {
    Type       ErrorType
    Message    string          // ユーザー向けメッセージ
    Cause      error           // 内部エラー
    Retryable  bool
    RetryAfter time.Duration
}

func (e *AppError) Error() string {
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Cause
}

// WrapAPIError はGitHub APIエラーをAppErrorに変換
func WrapAPIError(err error) *AppError {
    if err == nil {
        return nil
    }

    var ghErr *github.ErrorResponse
    if errors.As(err, &ghErr) {
        switch ghErr.Response.StatusCode {
        case 401:
            return &AppError{Type: ErrTypeAuth, Message: "認証に失敗しました", Retryable: false, Cause: err}
        case 403:
            if ghErr.Response.Header.Get("X-RateLimit-Remaining") == "0" {
                return &AppError{Type: ErrTypeRateLimit, Message: "レート制限に達しました", Retryable: true, Cause: err}
            }
            return &AppError{Type: ErrTypeAuth, Message: "アクセスが拒否されました", Retryable: false, Cause: err}
        case 404:
            return &AppError{Type: ErrTypeNotFound, Message: "リソースが見つかりません", Retryable: false, Cause: err}
        case 429:
            return &AppError{Type: ErrTypeRateLimit, Message: "リクエストが多すぎます", Retryable: true, Cause: err}
        default:
            if ghErr.Response.StatusCode >= 500 {
                return &AppError{Type: ErrTypeServer, Message: "GitHubサーバーエラー", Retryable: true, Cause: err}
            }
        }
    }

    return &AppError{Type: ErrTypeUnknown, Message: "予期しないエラー", Retryable: false, Cause: err}
}

// IsRetryable はエラーが再試行可能か判定
func IsRetryable(err error) bool {
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr.Retryable
    }
    return false
}
```

###github/interface.go（依存性注入用）

```go
package github

// Client はGitHub API操作のインターフェース
// テストだけでなく本番コードでも使用し、依存性注入を可能にする
type Client interface {
    // ワークフロー
    ListWorkflows(ctx context.Context, repo Repository) ([]Workflow, error)

    // 実行（Runs）
    ListRuns(ctx context.Context, repo Repository, opts *ListRunsOpts) ([]Run, error)
    CancelRun(ctx context.Context, repo Repository, runID int64) error
    RerunWorkflow(ctx context.Context, repo Repository, runID int64) error
    RerunFailedJobs(ctx context.Context, repo Repository, runID int64) error
    TriggerWorkflow(ctx context.Context, repo Repository, workflowFile, ref string, inputs map[string]interface{}) error

    // ジョブ
    ListJobs(ctx context.Context, repo Repository, runID int64) ([]Job, error)

    // ログ
    GetJobLogs(ctx context.Context, repo Repository, jobID int64) (string, error)

    // レート制限
    RateLimitRemaining() int
}
```

###github/client.go（実装）

```go
package github

import (
    "context"

    "github.com/google/go-github/v68/github"
    "golang.org/x/oauth2"
)

// Client はGitHub APIクライアント
type Client struct {
    gh *github.Client
}

// NewClient は新しいクライアントを作成
func NewClient(token string) *Client {
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    tc := oauth2.NewClient(context.Background(), ts)
    return &Client{gh: github.NewClient(tc)}
}

// === ワークフロー ===

// ListWorkflows はワークフロー一覧を取得
func (c *Client) ListWorkflows(ctx context.Context, owner, repo string) ([]Workflow, error)

// === 実行（Runs） ===

// ListRuns は実行一覧を取得
func (c *Client) ListRuns(ctx context.Context, owner, repo string, opts *ListRunsOpts) ([]Run, error)

// CancelRun は実行をキャンセル
func (c *Client) CancelRun(ctx context.Context, owner, repo string, runID int64) error

// RerunWorkflow は再実行
func (c *Client) RerunWorkflow(ctx context.Context, owner, repo string, runID int64) error

// RerunFailedJobs は失敗ジョブのみ再実行
func (c *Client) RerunFailedJobs(ctx context.Context, owner, repo string, runID int64) error

// TriggerWorkflow はworkflow_dispatchでトリガー
func (c *Client) TriggerWorkflow(ctx context.Context, owner, repo, workflowFile, ref string, inputs map[string]interface{}) error

// === ジョブ（★差別化機能） ===

// ListJobs はジョブ一覧を取得
func (c *Client) ListJobs(ctx context.Context, owner, repo string, runID int64) ([]Job, error)

// === ログ（★差別化機能） ===

// GetJobLogs はジョブのログを取得
func (c *Client) GetJobLogs(ctx context.Context, owner, repo string, jobID int64) (string, error)

// ListRunsOpts は実行一覧取得オプション
type ListRunsOpts struct {
    Branch     string
    Event      string
    Status     string
    WorkflowID int64
    PerPage    int
}
```

---

## 汎用リストパターン（lazydocker参考）

### FilteredList[T]

lazydockerの`FilteredList[T]`パターンを参考に、フィルタリング可能なスレッドセーフなリストを実装。

```go
//app/list.go

// FilteredList はフィルタリング可能なジェネリックリスト
type FilteredList[T any] struct {
    mu          sync.RWMutex
    allItems    []T
    filtered    []T
    filter      string
    selectedIdx int
    matchFn     func(item T, filter string) bool
}

// NewFilteredList は新しいFilteredListを作成
func NewFilteredList[T any](matchFn func(T, string) bool) *FilteredList[T] {
    return &FilteredList[T]{
        allItems:    make([]T, 0),
        filtered:    make([]T, 0),
        matchFn:     matchFn,
        selectedIdx: 0,
    }
}

// SetItems はアイテムを設定
func (l *FilteredList[T]) SetItems(items []T) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.allItems = items
    l.applyFilter()
}

// SetFilter はフィルターを設定
func (l *FilteredList[T]) SetFilter(filter string) {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.filter = filter
    l.applyFilter()
}

func (l *FilteredList[T]) applyFilter() {
    if l.filter == "" {
        l.filtered = l.allItems
        return
    }
    l.filtered = make([]T, 0)
    for _, item := range l.allItems {
        if l.matchFn(item, l.filter) {
            l.filtered = append(l.filtered, item)
        }
    }
    // 選択インデックスを範囲内に収める
    if l.selectedIdx >= len(l.filtered) {
        l.selectedIdx = max(0, len(l.filtered)-1)
    }
}

// Items はフィルタリング後のアイテムを返す
func (l *FilteredList[T]) Items() []T {
    l.mu.RLock()
    defer l.mu.RUnlock()
    return l.filtered
}

// Selected は選択中のアイテムを返す
func (l *FilteredList[T]) Selected() (T, bool) {
    l.mu.RLock()
    defer l.mu.RUnlock()
    if len(l.filtered) == 0 {
        var zero T
        return zero, false
    }
    return l.filtered[l.selectedIdx], true
}

// SelectNext は次のアイテムを選択
func (l *FilteredList[T]) SelectNext() {
    l.mu.Lock()
    defer l.mu.Unlock()
    if l.selectedIdx < len(l.filtered)-1 {
        l.selectedIdx++
    }
}

// SelectPrev は前のアイテムを選択
func (l *FilteredList[T]) SelectPrev() {
    l.mu.Lock()
    defer l.mu.Unlock()
    if l.selectedIdx > 0 {
        l.selectedIdx--
    }
}

// SelectedIndex は選択インデックスを返す
func (l *FilteredList[T]) SelectedIndex() int {
    l.mu.RLock()
    defer l.mu.RUnlock()
    return l.selectedIdx
}
```

### 使用例

```go
// Workflow用FilteredList
workflowList := NewFilteredList(func(w github.Workflow, filter string) bool {
    return strings.Contains(strings.ToLower(w.Name), strings.ToLower(filter))
})

// Run用FilteredList
runList := NewFilteredList(func(r github.Run, filter string) bool {
    return strings.Contains(r.Branch, filter) ||
           strings.Contains(r.Actor, filter)
})
```

---

## UIアーキテクチャ

### BubbleTeaモデル

```go
//app/app.go
package app

type Pane int

const (
    WorkflowsPane Pane = iota
    RunsPane
    LogsPane
)

// App はメインアプリケーション
type App struct {
    // データ（FilteredListパターン使用）
    repo      github.Repository
    workflows *FilteredList[github.Workflow]
    runs      *FilteredList[github.Run]
    jobs      *FilteredList[github.Job]

    // UI状態
    focusedPane Pane
    width       int
    height      int
    logView     *LogViewport  // Autoscroll付きログビュー

    // ポーリング
    logPoller *TickerTask

    // 状態
    loading bool
    err     error

    // ポップアップ
    showHelp    bool
    showConfirm bool
    confirmMsg  string
    confirmFn   func() tea.Cmd

    // フィルター（/キーで起動）
    filtering   bool
    filterInput textinput.Model

    // 依存
    client *github.Client
    keys   KeyMap
}
```

### 画面レイアウト（lazygitスタイル）

lazygit/lazydockerの3ペイン構成を採用。各ペインは独立したボーダーを持ち、フォーカス中は緑色で強調される。

#### 基本レイアウト（Logsペインにフォーカス中）

```
╭─ Workflows ────────╮╭─ Runs ──────────────────╮╭─ Logs ─────────────────────────────────╮
│                    ││                         ││                                        │
│ > ci.yml           ││ > #1234 ● main    2m ago││  Job: build (in_progress)              │
│   deploy.yml       ││   #1233 ✓ main    5m ago││  ────────────────────────────────────  │
│   test.yml         ││   #1232 ✗ feat/x 10m ago││  ✓ Set up job                     1s   │
│   release.yml      ││   #1231 ✓ main   15m ago││  ✓ Checkout                       2s   │
│                    ││   #1230 ⊘ main   20m ago││  ● Run tests                [running]  │
│                    ││                         ││    > go test -v ./...                  │
│                    ││                         ││    === RUN   TestClient                │
│                    ││                         ││    --- PASS: TestClient (0.02s)        │
│                    ││                         ││    === RUN   TestPolling               │
│                    ││                    2/5  ││    ▌                                   │
╰────────────────────╯╰─────────────────────────╯╰────────────────────────────────────────╯
 [c]ancel [r]erun [R]erun-failed [y]ank [L]fullscreen [?]help [q]uit
```

**凡例:**
- `╭──╮` 緑ボーダー = フォーカス中のペイン
- `╭──╮` グレーボーダー = 非フォーカスのペイン
- `>` = 選択中のアイテム（青背景でハイライト）
- `●` 黄 = 実行中、`✓` 緑 = 成功、`✗` 赤 = 失敗、`⊘` オレンジ = キャンセル済
- `2/5` = スクロール位置（現在位置/全件数）
- `▌` = ログのカーソル位置（Autoscroll時は最下部）

#### ペイン構成

```
┌────────────────────────────────────────────────────────────────────┐
│                          Terminal Width                            │
├──────────────┬──────────────────┬──────────────────────────────────┤
│  Workflows   │      Runs        │              Logs                │
│    (20%)     │      (25%)       │             (55%)                │
│              │                  │                                  │
│  ワークフロー │   実行一覧        │  選択中Runのログ                  │
│  一覧        │   (選択WFの)      │  (リアルタイム更新)               │
├──────────────┴──────────────────┴──────────────────────────────────┤
│                         Status Bar                                 │
│            コンテキスト依存のキーバインドヒント                        │
└────────────────────────────────────────────────────────────────────┘
```

#### 各ペインの詳細表示

**Workflowsペイン:**
```
╭─ Workflows ────────╮
│ > ci.yml           │  ← 選択中（青背景）
│   deploy.yml       │
│   test.yml         │
│   release.yml      │
│                    │
│                    │
│               1/4  │  ← スクロール位置
╰────────────────────╯
```

**Runsペイン:**
```
╭─ Runs ⣾ ───────────────────╮  ← ⣾ ローディングスピナー
│ > #1234 ● main      2m ago │  ← ● 黄=実行中
│   #1233 ✓ main      5m ago │  ← ✓ 緑=成功
│   #1232 ✗ feat/x   10m ago │  ← ✗ 赤=失敗
│   #1231 ✓ main     15m ago │
│   #1230 ⊘ main     20m ago │  ← ⊘ オレンジ=キャンセル
│                            │
│                       2/5  │
╰────────────────────────────╯
```

**Logsペイン（ステップ展開表示）:**
```
╭─ Logs ─────────────────────────────────────────╮
│  Run #1234 - CI (in_progress)                  │
│  Branch: main  Event: push  Actor: @user       │
│  ─────────────────────────────────────────     │
│  Job: build                                    │
│  ├─ ✓ Set up job                          1s  │
│  ├─ ✓ Checkout                            2s  │
│  ├─ ● Run tests                    [running]  │  ← 実行中ステップ
│  │    > go test -v ./...                       │
│  │    === RUN   TestClient                     │
│  │    --- PASS: TestClient (0.02s)             │
│  │    === RUN   TestPolling                    │
│  │    ▌                                        │  ← カーソル
│  ├─ ○ Build                         [pending]  │
│  └─ ○ Upload artifacts              [pending]  │
╰────────────────────────────────────────────────╯
```

**Logsペイン（失敗時 - 自動ジャンプ）:**
```
╭─ Logs ─────────────────────────────────────────╮
│  Run #1232 - CI (failure)                      │
│  Branch: feat/x  Event: push  Actor: @user     │
│  ─────────────────────────────────────────     │
│  Job: test                                     │
│  ├─ ✓ Set up job                          1s  │
│  ├─ ✓ Checkout                            2s  │
│  ├─ ✗ Run tests                          45s  │  ← 失敗ステップ（赤）
│  │    > go test -v ./...                       │
│  │    === RUN   TestFoo                        │
│  │    --- FAIL: TestFoo (0.01s)                │  ← 自動ジャンプ位置
│  │        foo_test.go:42: expected 1, got 2    │
│  │    FAIL                                     │
│  │    exit status 1                            │
│  └─ - Post Checkout                   [skipped]│
╰────────────────────────────────────────────────╯
```

#### ポップアップ表示

**確認ダイアログ（cキーでキャンセル時）:**
```
╭─ Workflows ───╮╭─ Runs ─────────╮╭─ Logs ────────────────╮
│               ││                ││                       │
│ > ci.yml      ││ > #1234 ● 2m   ││                       │
│   deploy.yml  ││   #1233 ✓ 5m   ││  ╭─────────────────╮  │
│               ││   #1232 ✗ 10m  ││  │ Cancel run      │  │
│               ││                ││  │ #1234?          │  │
│               ││                ││  │                 │  │
│               ││                ││  │ [y]es  [n]o     │  │
│               ││                ││  ╰─────────────────╯  │
│               ││                ││                       │
╰───────────────╯╰────────────────╯╰───────────────────────╯
```

**ヘルプポップアップ（?キー）:**
```
                    ╭─────────────────────────────────────╮
                    │          lazyactions help           │
                    ├─────────────────────────────────────┤
                    │  Navigation                         │
                    │  ─────────────────────────────────  │
                    │  j/↓       Down                     │
                    │  k/↑       Up                       │
                    │  h/←       Previous pane            │
                    │  l/→       Next pane                │
                    │  Tab       Next pane                │
                    │  S-Tab     Previous pane            │
                    │                                     │
                    │  Actions                            │
                    │  ─────────────────────────────────  │
                    │  t         Trigger workflow         │
                    │  c         Cancel run               │
                    │  r         Rerun workflow           │
                    │  R         Rerun failed jobs only   │
                    │  y         Copy URL to clipboard    │
                    │                                     │
                    │  View                               │
                    │  ─────────────────────────────────  │
                    │  /         Filter                   │
                    │  L         Full-screen log          │
                    │  Enter     Expand/Collapse          │
                    │  Esc       Back/Close               │
                    │  ?         Toggle this help         │
                    │  q         Quit                     │
                    ╰─────────────────────────────────────╯
```

**フィルターモード（/キー）:**
```
╭─ Workflows ───╮╭─ Runs ─────────╮╭─ Logs ────────────────╮
│               ││                ││                       │
│ > ci.yml      ││ > #1234 ● 2m   ││                       │
│               ││   #1233 ✓ 5m   ││                       │
│               ││                ││                       │
╰───────────────╯╰────────────────╯╰───────────────────────╯
 Filter: main▌                                    [Esc] cancel
```

**ログ全画面モード（Lキー）:**
```
╭─ Logs ─ Run #1234 - CI ─────────────────────────────────────────────╮
│                                                                      │
│  Job: build (in_progress)                                           │
│  ───────────────────────────────────────────────────────────────    │
│  ✓ Set up job                                                  1s   │
│  ✓ Checkout                                                    2s   │
│  ● Run tests                                            [running]   │
│    > go test -v ./...                                                │
│    === RUN   TestClient                                              │
│    --- PASS: TestClient (0.02s)                                      │
│    === RUN   TestPolling                                             │
│    --- PASS: TestPolling (0.05s)                                     │
│    === RUN   TestAuth                                                │
│    --- PASS: TestAuth (0.01s)                                        │
│    PASS                                                              │
│    ok      github.com/owner/repo/github  0.08s              │
│    ▌                                                                 │
│                                                                      │
│                                                                      │
╰──────────────────────────────────────────────────────────────────────╯
 [y]ank [Esc]back [?]help [q]uit
```

### Lazy-styleスタイリング詳細

lazygit/lazydockerの視覚的特徴を忠実に再現する。

#### 1. フォーカス表示（最重要）

lazygitの特徴: **フォーカス中のパネルは緑ボーダー、タイトルも緑背景**

```go
//app/styles.go

var (
    // lazygit標準色
    FocusedColor   = lipgloss.Color("#00FF00")  // 緑
    UnfocusedColor = lipgloss.Color("#666666")  // グレー

    // パネルスタイル
    FocusedPane = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(FocusedColor)

    UnfocusedPane = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(UnfocusedColor)

    // タイトル（パネル上部に表示）
    FocusedTitle = lipgloss.NewStyle().
        Background(FocusedColor).
        Foreground(lipgloss.Color("#000000")).
        Padding(0, 1).Bold(true)

    UnfocusedTitle = lipgloss.NewStyle().
        Foreground(UnfocusedColor).
        Padding(0, 1)
)
```

**レイアウト図（フォーカス表示付き）:**
```
┌ Workflows ─┐┌─ Runs ──────┐┌─ Logs ──────────────────────┐
│            ││             ││  ← 緑ボーダー（フォーカス中）│
│ > ci.yml   ││ > #123 ● 2m ││                              │
│   deploy   ││   #122 ✓ 5m ││  Job: build                  │
└────────────┘└─────────────┘└──────────────────────────────┘
  ↑グレーボーダー（非フォーカス）
```

#### 2. ステータスアイコン・カラー

```go
var (
    SuccessStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
    FailureStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
    RunningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
    QueuedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
    CancelledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8800"))
)

// StatusIcon はステータスに応じたアイコンを返す
func StatusIcon(status, conclusion string) string {
    switch {
    case status == "in_progress":
        return RunningStyle.Render("●")
    case status == "queued":
        return QueuedStyle.Render("○")
    case conclusion == "success":
        return SuccessStyle.Render("✓")
    case conclusion == "failure":
        return FailureStyle.Render("✗")
    case conclusion == "cancelled":
        return CancelledStyle.Render("⊘")
    default:
        return " "
    }
}
```

#### 3. 選択行のハイライト

```go
// 選択中: 白文字＋青背景（lazygit標準）
SelectedItem = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FFFFFF")).
    Background(lipgloss.Color("#0055AA"))

// 非選択: 薄いグレー
NormalItem = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#AAAAAA"))

func RenderItem(text string, selected bool) string {
    if selected {
        return SelectedItem.Render("> " + text)
    }
    return NormalItem.Render("  " + text)
}
```

#### 4. 確認ダイアログ（破壊的操作用）

```go
// オレンジボーダーで警告を表現
ConfirmDialog = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#FF8800")).
    Padding(1, 2).
    Width(50)

func (a *App) renderConfirmDialog() string {
    content := lipgloss.JoinVertical(lipgloss.Center,
        lipgloss.NewStyle().Bold(true).Render(a.confirmMsg),
        "",
        "[y] Yes  [n] No  [Esc] Cancel",
    )
    dialog := ConfirmDialog.Render(content)
    return lipgloss.Place(a.width, a.height,
        lipgloss.Center, lipgloss.Center, dialog)
}
```

**確認ダイアログ例:**
```
           ┌──────────────────────────────────────┐
           │                                      │
           │   Cancel workflow run #123?          │
           │                                      │
           │        [y] Yes  [n] No               │
           │                                      │
           └──────────────────────────────────────┘
```

#### 5. ヘルプポップアップ（?キー）

```go
HelpPopup = lipgloss.NewStyle().
    Border(lipgloss.DoubleBorder()).
    BorderForeground(lipgloss.Color("#00FFFF")).
    Padding(1, 2)

func (a *App) renderHelp() string {
    help := `
┌─ Navigation ──────────────────────┐
│ j/↓         Move down             │
│ k/↑         Move up               │
│ h/←/Tab     Previous pane         │
│ l/→/S-Tab   Next pane             │
│ Enter       Expand/Select         │
├─ Actions ─────────────────────────┤
│ t           Trigger workflow      │
│ c           Cancel run            │
│ r           Rerun workflow        │
│ R           Rerun failed jobs     │
│ y           Copy URL/log          │
├─ View ────────────────────────────┤
│ /           Filter                │
│ L           Full-screen log       │
│ Esc         Close/Back            │
├─ General ─────────────────────────┤
│ ?           Toggle help           │
│ q           Quit                  │
└───────────────────────────────────┘
`
    return lipgloss.Place(a.width, a.height,
        lipgloss.Center, lipgloss.Center,
        HelpPopup.Render(help))
}
```

#### 6. ローディングスピナー

```go
import "github.com/charmbracelet/bubbles/spinner"

// 初期化
a.spinner = spinner.New()
a.spinner.Spinner = spinner.Dot  // ⣾⣽⣻⢿⡿⣟⣯⣷
a.spinner.Style = RunningStyle

// パネルタイトルに表示
func (a *App) runsTitle() string {
    if a.loading {
        return fmt.Sprintf("Runs %s", a.spinner.View())
    }
    return "Runs"
}
```

#### 7. ステータスバー（コンテキスト依存）

```go
StatusBar = lipgloss.NewStyle().
    Background(lipgloss.Color("#333333")).
    Padding(0, 1)

func (a *App) renderStatusBar() string {
    // フォーカス中のペインに応じてヒントを変更
    hints := map[Pane]string{
        WorkflowsPane: "[t]rigger [/]filter [?]help [q]uit",
        RunsPane:      "[c]ancel [r]erun [R]erun-failed [y]ank [?]help [q]uit",
        LogsPane:      "[L]fullscreen [y]ank [Esc]back [?]help [q]uit",
    }

    // エラーがあれば赤で表示
    if a.err != nil {
        return StatusBar.
            Foreground(lipgloss.Color("#FF0000")).
            Width(a.width).
            Render("Error: " + a.err.Error())
    }

    return StatusBar.Width(a.width).Render(hints[a.focusedPane])
}
```

#### 8. スクロール位置表示

```go
// リストの右下に "3/10" のように表示
func ScrollPosition(current, total int) string {
    return lipgloss.NewStyle().
        Foreground(lipgloss.Color("#666666")).
        Render(fmt.Sprintf("%d/%d", current+1, total))
}
```

### キーバインド

```go
//app/keymap.go
type KeyMap struct {
    // ナビゲーション
    Up        key.Binding  // k, ↑
    Down      key.Binding  // j, ↓
    Left      key.Binding  // h, ←（前のペインへ）
    Right     key.Binding  // l, →（次のペインへ）
    Tab       key.Binding  // Tab（次のペインへ）
    ShiftTab  key.Binding  // Shift+Tab（前のペインへ）
    Enter     key.Binding  // Enter（選択/展開）

    // アクション
    Trigger   key.Binding  // t: workflow_dispatch
    Cancel    key.Binding  // c: キャンセル（確認あり）
    Rerun     key.Binding  // r: 再実行
    Yank      key.Binding  // y: クリップボードコピー
    Filter    key.Binding  // /: フィルター
    Refresh   key.Binding  // R: 更新
    FullLog   key.Binding  // L: ログ全画面

    // UI
    Help      key.Binding  // ?: ヘルプポップアップ
    Quit      key.Binding  // q: 終了
    Escape    key.Binding  // Esc: 戻る/閉じる
}
```

---

## BubbleTeaメッセージ・コマンド設計

###app/messages.go（明示的なメッセージ型）

BubbleTeaのMVUパターンでは、状態遷移を明示的なメッセージ型で表現する。

```go
package app

// === データ取得結果 ===

// WorkflowsLoadedMsg はワークフロー取得完了を示す
type WorkflowsLoadedMsg struct {
    Workflows []github.Workflow
    Err       error
}

// RunsLoadedMsg は実行一覧取得完了を示す
type RunsLoadedMsg struct {
    Runs []github.Run
    Err  error
}

// JobsLoadedMsg はジョブ取得完了を示す
type JobsLoadedMsg struct {
    Jobs []github.Job
    Err  error
}

// LogsLoadedMsg はログ取得完了を示す
type LogsLoadedMsg struct {
    Logs string
    Err  error
}

// === アクション結果 ===

// RunCancelledMsg はキャンセル完了を示す
type RunCancelledMsg struct {
    RunID int64
    Err   error
}

// RunRerunMsg は再実行完了を示す
type RunRerunMsg struct {
    RunID int64
    Err   error
}

// WorkflowTriggeredMsg はworkflow_dispatch完了を示す
type WorkflowTriggeredMsg struct {
    Workflow string
    Err      error
}

// === UI状態 ===

// FlashMsg は一時メッセージ表示を示す
type FlashMsg struct {
    Message  string
    Duration time.Duration
}

// FlashClearMsg は一時メッセージクリアを示す
type FlashClearMsg struct{}

// TickMsg はポーリング継続を示す
type TickMsg struct {
    Time time.Time
}
```

###app/commands.go（不変値キャプチャ）

クロージャでの競合状態を防ぐため、tea.Cmd生成時に必要な値をコピーする。

```go
package app

// fetchWorkflows はワークフロー取得コマンドを生成
// ★ 不変値をキャプチャしてクロージャに渡す
func (a *App) fetchWorkflows() tea.Cmd {
    // 値をコピー（競合状態防止）
    client := a.client
    repo := a.repo

    return func() tea.Msg {
        workflows, err := client.ListWorkflows(context.Background(), repo)
        return WorkflowsLoadedMsg{Workflows: workflows, Err: err}
    }
}

// fetchRuns は実行一覧取得コマンドを生成
func (a *App) fetchRuns(workflowID int64) tea.Cmd {
    client := a.client
    repo := a.repo

    return func() tea.Msg {
        runs, err := client.ListRuns(context.Background(), repo, &github.ListRunsOpts{
            WorkflowID: workflowID,
            PerPage:    20,
        })
        return RunsLoadedMsg{Runs: runs, Err: err}
    }
}

// fetchJobs はジョブ取得コマンドを生成
func (a *App) fetchJobs(runID int64) tea.Cmd {
    client := a.client
    repo := a.repo

    return func() tea.Msg {
        jobs, err := client.ListJobs(context.Background(), repo, runID)
        return JobsLoadedMsg{Jobs: jobs, Err: err}
    }
}

// fetchLogs はログ取得コマンドを生成
func (a *App) fetchLogs(jobID int64) tea.Cmd {
    client := a.client
    repo := a.repo

    return func() tea.Msg {
        logs, err := client.GetJobLogs(context.Background(), repo, jobID)
        return LogsLoadedMsg{Logs: logs, Err: err}
    }
}

// cancelRun はキャンセルコマンドを生成
func (a *App) cancelRun(runID int64) tea.Cmd {
    client := a.client
    repo := a.repo

    return func() tea.Msg {
        err := client.CancelRun(context.Background(), repo, runID)
        return RunCancelledMsg{RunID: runID, Err: err}
    }
}

// showFlash は一時メッセージ表示コマンドを生成
func showFlash(msg string, duration time.Duration) tea.Cmd {
    return func() tea.Msg {
        return FlashMsg{Message: msg, Duration: duration}
    }
}

// clearFlashAfter は一定時間後にFlashをクリアするコマンド
func clearFlashAfter(d time.Duration) tea.Cmd {
    return tea.Tick(d, func(t time.Time) tea.Msg {
        return FlashClearMsg{}
    })
}
```

---

## ログストリーミング実装（★最重要機能）

### アーキテクチャ

```
選択変更時:
  Run選択 → Jobs取得 → 最初のJob選択 → Logs取得 → 表示

ポーリング（実行中のみ）:
  2-3秒間隔 → GetJobLogs → 差分更新 → 表示更新
```

### TickerTaskパターン（lazydocker参考）

lazydockerの`NewTickerTask`パターンを参考に、コンテキストベースのキャンセレーションでポーリングを制御する。

```go
//app/ticker.go

// TickerTask は定期的なポーリングタスクを管理
type TickerTask struct {
    ctx      context.Context
    cancel   context.CancelFunc
    interval time.Duration
    taskFn   func(ctx context.Context) tea.Msg
}

// NewTickerTask は新しいTickerTaskを作成
func NewTickerTask(interval time.Duration, fn func(ctx context.Context) tea.Msg) *TickerTask {
    ctx, cancel := context.WithCancel(context.Background())
    return &TickerTask{
        ctx:      ctx,
        cancel:   cancel,
        interval: interval,
        taskFn:   fn,
    }
}

// Start はポーリングを開始
func (t *TickerTask) Start() tea.Cmd {
    return func() tea.Msg {
        ticker := time.NewTicker(t.interval)
        defer ticker.Stop()

        for {
            select {
            case <-t.ctx.Done():
                return nil
            case <-ticker.C:
                if msg := t.taskFn(t.ctx); msg != nil {
                    return msg
                }
            }
        }
    }
}

// Stop はポーリングを停止
func (t *TickerTask) Stop() {
    t.cancel()
}
```

### Autoscroll付きログビューポート

lazydockerのAutoscroll機能を参考に、新しいログが追加されたら自動的に最下部へスクロール。

```go
//app/logview.go

// LogViewport はAutoscroll機能付きのログビューポート
type LogViewport struct {
    viewport   viewport.Model
    autoscroll bool  // 最下部にいる場合は自動スクロール
}

// NewLogViewport は新しいLogViewportを作成
func NewLogViewport(width, height int) *LogViewport {
    vp := viewport.New(width, height)
    return &LogViewport{
        viewport:   vp,
        autoscroll: true,
    }
}

// SetContent はログコンテンツを設定
func (l *LogViewport) SetContent(content string) {
    wasAtBottom := l.isAtBottom()
    l.viewport.SetContent(content)

    // Autoscrollが有効で、以前最下部にいた場合は最下部へ移動
    if l.autoscroll && wasAtBottom {
        l.viewport.GotoBottom()
    }
}

func (l *LogViewport) isAtBottom() bool {
    return l.viewport.AtBottom()
}

// Update はビューポートの更新を処理
func (l *LogViewport) Update(msg tea.Msg) (*LogViewport, tea.Cmd) {
    var cmd tea.Cmd
    l.viewport, cmd = l.viewport.Update(msg)

    // ユーザーが手動でスクロールしたらAutoscrollを無効化
    // 最下部に戻ったら再度有効化
    l.autoscroll = l.isAtBottom()

    return l, cmd
}
```

### 実装

```go
// ポーリングメッセージ
type pollLogsMsg struct{}
type logsUpdatedMsg struct {
    logs string
    err  error
}

// App構造体にTickerTaskを追加
type App struct {
    // ... 既存フィールド ...
    logPoller  *TickerTask
    logView    *LogViewport
}

// ポーリング開始（TickerTaskパターン使用）
func (a *App) startLogPolling() tea.Cmd {
    // 既存のポーラーがあれば停止
    if a.logPoller != nil {
        a.logPoller.Stop()
    }

    a.logPoller = NewTickerTask(2*time.Second, func(ctx context.Context) tea.Msg {
        if len(a.jobs) == 0 {
            return nil
        }
        job := a.jobs[a.selectedJob]
        logs, err := a.client.GetJobLogs(ctx, a.repo.Owner, a.repo.Name, job.ID)
        if ctx.Err() != nil {
            return nil // キャンセルされた
        }
        return logsUpdatedMsg{logs: logs, err: err}
    })

    return a.logPoller.Start()
}

// ポーリング停止
func (a *App) stopLogPolling() {
    if a.logPoller != nil {
        a.logPoller.Stop()
        a.logPoller = nil
    }
}

// Update
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case logsUpdatedMsg:
        if msg.err == nil {
            a.logView.SetContent(msg.logs)
            // 失敗ステップがあれば自動ジャンプ
            if step := a.findFailedStep(); step != nil {
                a.jumpToStep(step)
            }
        }
        // 実行中なら次のポーリングをスケジュール
        if a.isSelectedRunRunning() {
            return a, a.startLogPolling()
        }
    }
    return a, nil
}
```

### 適応型ポーリング（レート制限対応）

GitHub APIのレート制限を考慮して、残りリクエスト数に応じてポーリング間隔を調整する。

```go
//app/polling.go

// AdaptivePoller はレート制限に応じて間隔を調整するポーラー
type AdaptivePoller struct {
    baseInterval time.Duration
    maxInterval  time.Duration
    client       github.Client
}

// NewAdaptivePoller は新しいAdaptivePollerを作成
func NewAdaptivePoller(client github.Client) *AdaptivePoller {
    return &AdaptivePoller{
        baseInterval: 2 * time.Second,
        maxInterval:  30 * time.Second,
        client:       client,
    }
}

// NextInterval はレート制限に基づいて次のポーリング間隔を計算
func (p *AdaptivePoller) NextInterval() time.Duration {
    remaining := p.client.RateLimitRemaining()

    switch {
    case remaining < 100:
        // 残り少ない: 最大間隔
        return p.maxInterval
    case remaining < 500:
        // 注意レベル: 間隔を2倍
        return p.baseInterval * 2
    case remaining < 1000:
        // 軽度注意: 間隔を1.5倍
        return time.Duration(float64(p.baseInterval) * 1.5)
    default:
        // 通常: 基本間隔
        return p.baseInterval
    }
}

// StartAdaptivePolling は適応型ポーリングを開始
func (a *App) startAdaptivePolling() tea.Cmd {
    if a.logPoller != nil {
        a.logPoller.Stop()
    }

    interval := a.adaptivePoller.NextInterval()

    a.logPoller = NewTickerTask(interval, func(ctx context.Context) tea.Msg {
        job, ok := a.jobs.Selected()
        if !ok {
            return nil
        }
        logs, err := a.client.GetJobLogs(ctx, a.repo, job.ID)
        if ctx.Err() != nil {
            return nil
        }
        return LogsLoadedMsg{Logs: logs, Err: err}
    })

    return a.logPoller.Start()
}
```

**レート制限ステータス表示:**

```go
// ステータスバーにレート制限情報を表示
func (a *App) renderRateLimitStatus() string {
    remaining := a.client.RateLimitRemaining()

    var style lipgloss.Style
    switch {
    case remaining < 100:
        style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
    case remaining < 500:
        style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
    default:
        style = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
    }

    return style.Render(fmt.Sprintf("API: %d/5000", remaining))
}
```

---

### 失敗ステップ自動ジャンプ

```go
func (a *App) findFailedStep() *github.Step {
    if len(a.jobs) == 0 {
        return nil
    }
    for _, step := range a.jobs[a.selectedJob].Steps {
        if step.IsFailed() {
            return &step
        }
    }
    return nil
}

func (a *App) jumpToStep(step *github.Step) {
    lines := strings.Split(a.logs, "\n")
    for i, line := range lines {
        if strings.Contains(line, step.Name) {
            a.viewport.SetYOffset(i)
            return
        }
    }
}
```

---

## テスト駆動開発（TDD）

### テスト構成

```
├── app/
│   ├── app.go
│   └── app_test.go      # UIロジックのテスト
├── github/
│   ├── client.go
│   ├── client_test.go   # APIクライアントのテスト
│   └── mock_test.go     # モック定義
├── auth/
│   ├── token.go
│   └── token_test.go    # 認証のテスト
└── repo/
    ├── detect.go
    └── detect_test.go   # リポジトリ検出のテスト
```

### テストカバレッジ目標

パッケージごとのカバレッジ目標:

| パッケージ | 目標 | 重点テスト項目 |
|-----------|------|---------------|
| `auth/` | 80% | トークン取得優先順位、SecureToken |
| `repo/` | 85% | URL解析（SSH/HTTPS）、バリデーション |
| `github/` | 75% | API呼び出し、エラーハンドリング |
| `app/` | 70% | キー入力、状態遷移、UI更新 |
| **全体** | **70%** | - |

### モック戦略（呼び出し追跡付き）

```go
//github/mock.go（★本番コードに配置）

// MockClient はテスト用モック（呼び出し追跡機能付き）
type MockClient struct {
    // 戻り値設定
    Workflows []Workflow
    Runs      []Run
    Jobs      []Job
    Logs      string
    Err       error

    // 呼び出し追跡
    calls []MethodCall
    mu    sync.Mutex
}

// MethodCall はメソッド呼び出しを記録
type MethodCall struct {
    Method string
    Args   []interface{}
    Time   time.Time
}

// RecordCall は呼び出しを記録
func (m *MockClient) RecordCall(method string, args ...interface{}) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.calls = append(m.calls, MethodCall{
        Method: method,
        Args:   args,
        Time:   time.Now(),
    })
}

// Calls は記録された呼び出しを返す
func (m *MockClient) Calls() []MethodCall {
    m.mu.Lock()
    defer m.mu.Unlock()
    return append([]MethodCall{}, m.calls...)
}

// CallCount は特定メソッドの呼び出し回数を返す
func (m *MockClient) CallCount(method string) int {
    m.mu.Lock()
    defer m.mu.Unlock()
    count := 0
    for _, c := range m.calls {
        if c.Method == method {
            count++
        }
    }
    return count
}

// Reset は呼び出し履歴をクリア
func (m *MockClient) Reset() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.calls = nil
}

// インターフェース実装
func (m *MockClient) ListWorkflows(ctx context.Context, repo Repository) ([]Workflow, error) {
    m.RecordCall("ListWorkflows", repo)
    return m.Workflows, m.Err
}

func (m *MockClient) ListRuns(ctx context.Context, repo Repository, opts *ListRunsOpts) ([]Run, error) {
    m.RecordCall("ListRuns", repo, opts)
    return m.Runs, m.Err
}

func (m *MockClient) ListJobs(ctx context.Context, repo Repository, runID int64) ([]Job, error) {
    m.RecordCall("ListJobs", repo, runID)
    return m.Jobs, m.Err
}

func (m *MockClient) GetJobLogs(ctx context.Context, repo Repository, jobID int64) (string, error) {
    m.RecordCall("GetJobLogs", repo, jobID)
    return m.Logs, m.Err
}

func (m *MockClient) CancelRun(ctx context.Context, repo Repository, runID int64) error {
    m.RecordCall("CancelRun", repo, runID)
    return m.Err
}

func (m *MockClient) RerunWorkflow(ctx context.Context, repo Repository, runID int64) error {
    m.RecordCall("RerunWorkflow", repo, runID)
    return m.Err
}

func (m *MockClient) RerunFailedJobs(ctx context.Context, repo Repository, runID int64) error {
    m.RecordCall("RerunFailedJobs", repo, runID)
    return m.Err
}

func (m *MockClient) TriggerWorkflow(ctx context.Context, repo Repository, workflowFile, ref string, inputs map[string]interface{}) error {
    m.RecordCall("TriggerWorkflow", repo, workflowFile, ref, inputs)
    return m.Err
}

func (m *MockClient) RateLimitRemaining() int {
    m.RecordCall("RateLimitRemaining")
    return 5000 // デフォルトは十分な残量
}
```

### モック使用例

```go
//app/app_test.go

func TestApp_PollingBehavior(t *testing.T) {
    mock := &github.MockClient{
        Jobs: []github.Job{{ID: 1, Name: "build", Status: "in_progress"}},
        Logs: "Running tests...",
    }

    app := NewApp(mock, github.Repository{Owner: "o", Name: "r"})
    app.runs.SetItems([]github.Run{{ID: 100, Status: "in_progress"}})

    // 実行中のRunを選択
    app.focusedPane = RunsPane
    app.Update(tea.KeyMsg{Type: tea.KeyEnter})

    // ログ取得が呼ばれたことを確認
    assert.Equal(t, 1, mock.CallCount("GetJobLogs"))

    // ポーリング継続を確認
    time.Sleep(3 * time.Second)
    assert.GreaterOrEqual(t, mock.CallCount("GetJobLogs"), 2)
}

func TestApp_StopsPollingOnCompletion(t *testing.T) {
    mock := &github.MockClient{
        Jobs: []github.Job{{ID: 1, Name: "build", Status: "completed", Conclusion: "success"}},
        Logs: "Done!",
    }

    app := NewApp(mock, github.Repository{Owner: "o", Name: "r"})
    app.runs.SetItems([]github.Run{{ID: 100, Status: "completed", Conclusion: "success"}})

    // 完了したRunを選択
    app.focusedPane = RunsPane
    app.Update(tea.KeyMsg{Type: tea.KeyEnter})

    initialCalls := mock.CallCount("GetJobLogs")
    time.Sleep(3 * time.Second)

    // ポーリングが停止していることを確認
    assert.Equal(t, initialCalls, mock.CallCount("GetJobLogs"))
}
```

### テスト例

```go
//github/client_test.go
func TestClient_ListWorkflows(t *testing.T) {
    tests := []struct {
        name    string
        want    []Workflow
        wantErr bool
    }{
        {
            name: "returns workflows",
            want: []Workflow{
                {ID: 1, Name: "CI", Path: ".github/workflows/ci.yml"},
                {ID: 2, Name: "Deploy", Path: ".github/workflows/deploy.yml"},
            },
        },
        {
            name:    "handles empty",
            want:    []Workflow{},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // httptest でモックサーバー
            server := httptest.NewServer(/* ... */)
            defer server.Close()

            client := NewClientWithURL(server.URL, "token")
            got, err := client.ListWorkflows(context.Background(), "owner", "repo")

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}

//app/app_test.go
func TestApp_SelectRun(t *testing.T) {
    mock := &MockClient{
        Jobs: []Job{{ID: 1, Name: "build"}},
        Logs: "=== RUN TestFoo\n--- PASS: TestFoo",
    }

    app := NewApp(mock, Repository{Owner: "o", Name: "r"})
    app.runs = []Run{{ID: 100, Status: "completed"}}

    // Run選択時にJobsとLogsが取得される
    app.selectedRun = 0
    cmd := app.onRunSelected()

    // コマンド実行
    msg := cmd()

    // 結果確認
    assert.Len(t, app.jobs, 1)
    assert.Contains(t, app.logs, "TestFoo")
}
```

---

## 実装順序

### Week 1: 基盤

| Day | タスク | TDD |
|-----|-------|-----|
| 1 | プロジェクト初期化、go.mod、Makefile | - |
| 2 | auth/: トークン取得 | token_test.go → token.go |
| 3 | repo/: リポジトリ検出 | detect_test.go → detect.go |
| 4-5 | github/: 基本API | client_test.go → client.go |

### Week 2: コア機能

| Day | タスク | TDD |
|-----|-------|-----|
| 1 | github: ListJobs, GetJobLogs | client_test.go → client.go |
| 2-3 | app: 3ペインレイアウト | app_test.go → app.go |
| 4-5 | app: ログストリーミング | app_test.go → app.go |

### Week 3: 機能完成

| Day | タスク |
|-----|-------|
| 1 | キャンセル・再実行（確認ダイアログ） |
| 2 | フィルター、クリップボードコピー |
| 3 | ヘルプポップアップ、ステータスバー |
| 4-5 | テスト追加、リファクタリング、README |

---

## Makefile

```makefile
.PHONY: test build run lint cover cover-check

# カバレッジ閾値
COVERAGE_THRESHOLD := 70

# テスト
test:
	go test -v -race ./...

# カバレッジ計測
cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# カバレッジ閾値チェック（CI用）
cover-check:
	@go test -race -coverprofile=coverage.out ./... > /dev/null 2>&1
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $${COVERAGE}%"; \
	if [ $$(echo "$${COVERAGE} < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "ERROR: Coverage $${COVERAGE}% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	fi; \
	echo "OK: Coverage meets threshold"

# パッケージ別カバレッジ
cover-by-pkg:
	@echo "=== Coverage by package ==="
	@go test -race -coverprofile=coverage.out ./... > /dev/null 2>&1
	@go tool cover -func=coverage.out | grep -E '^(github.com|total:)'

# ビルド
build:
	go build -o bin/lazyactions ./cmd/lazyactions

# 実行
run:
	go run ./cmd/lazyactions

# リント
lint:
	golangci-lint run

# CI用（リント + カバレッジチェック + ビルド）
ci: lint cover-check build

# 開発用（フォーマット + リント + テスト）
dev: fmt lint test

# フォーマット
fmt:
	go fmt ./...
	goimports -w .

# クリーン
clean:
	rm -rf bin/ coverage.out coverage.html
```

### CI設定例（GitHub Actions）

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Lint
        uses: golangci/golangci-lint-action@v4

      - name: Test with coverage
        run: make cover-check

      - name: Build
        run: make build
```

---

## クリップボードコピー（F09）

`y`キーでログ/URLをクリップボードにコピーする。

```go
//app/clipboard.go

import "github.com/atotto/clipboard"

// CopyToClipboard はテキストをクリップボードにコピー
func CopyToClipboard(text string) error {
    return clipboard.WriteAll(text)
}

// コピー対象の決定（フォーカス位置による）
func (a *App) getYankContent() string {
    switch a.focusedPane {
    case LogsPane:
        // 表示中のログ全体をコピー
        return a.logView.Content()
    case RunsPane:
        // 選択中RunのURLをコピー
        if run, ok := a.runs.Selected(); ok {
            return run.URL
        }
    case WorkflowsPane:
        // 選択中WorkflowのYAMLパスをコピー
        if wf, ok := a.workflows.Selected(); ok {
            return wf.Path
        }
    }
    return ""
}

// yキー押下時
func (a *App) handleYank() tea.Cmd {
    content := a.getYankContent()
    if content == "" {
        return nil
    }
    if err := CopyToClipboard(content); err != nil {
        a.err = err
        return nil
    }
    // ステータスバーに「Copied!」を一時表示
    return a.showFlashMessage("Copied to clipboard!")
}
```

---

## workflow_dispatch 実行（tキー）

`t`キーでworkflow_dispatchトリガー。inputs対応。

```go
//app/trigger.go

// TriggerState はworkflow_dispatch実行の状態
type TriggerState struct {
    workflow  github.Workflow
    inputs    []WorkflowInput   // ワークフローから取得したinputs定義
    values    map[string]string // ユーザー入力値
    focusIdx  int               // フォーカス中のinputインデックス
    ref       string            // ブランチ名
}

// WorkflowInput はworkflow_dispatchのinput定義
type WorkflowInput struct {
    Name        string
    Description string
    Required    bool
    Default     string
    Type        string // string, boolean, choice, environment
    Options     []string // Type=choice の場合
}

// トリガーダイアログのレンダリング
func (a *App) renderTriggerDialog() string {
    if a.triggerState == nil {
        return ""
    }

    var b strings.Builder
    b.WriteString(fmt.Sprintf("Trigger: %s\n", a.triggerState.workflow.Name))
    b.WriteString("─────────────────────────────\n")
    b.WriteString(fmt.Sprintf("Branch: %s\n\n", a.triggerState.ref))

    // 各inputを表示
    for i, input := range a.triggerState.inputs {
        prefix := "  "
        if i == a.triggerState.focusIdx {
            prefix = "> "
        }

        required := ""
        if input.Required {
            required = " *"
        }

        value := a.triggerState.values[input.Name]
        if value == "" {
            value = input.Default
        }

        b.WriteString(fmt.Sprintf("%s%s%s: %s\n", prefix, input.Name, required, value))
        if input.Description != "" && i == a.triggerState.focusIdx {
            b.WriteString(fmt.Sprintf("    %s\n", input.Description))
        }
    }

    b.WriteString("\n[Enter] Run  [Esc] Cancel")
    return ConfirmDialog.Render(b.String())
}
```

**トリガーダイアログ例:**
```
╭──────────────────────────────────────────╮
│ Trigger: deploy.yml                      │
│ ─────────────────────────────────        │
│ Branch: main                             │
│                                          │
│ > environment *: production              │
│     Target deployment environment        │
│   version: v1.2.3                        │
│   dry_run: false                         │
│                                          │
│ [Enter] Run  [Esc] Cancel                │
╰──────────────────────────────────────────╯
```

---

## 非機能要件

task.mdの非機能要件を実装に反映:

### パフォーマンス

| 要件 | 実装方針 |
|------|---------|
| 起動時間: 1秒以内 | 遅延読み込み（API呼び出しは起動後に非同期実行） |
| API レスポンス表示: 500ms以内 | ローディングスピナー表示、結果キャッシュ |
| ログストリーミング間隔: 2秒 | `TickerTask(2*time.Second)` |
| メモリ使用量: 100MB以下 | ログは最新N行のみ保持、古いログはGC |

```go
// 起動シーケンス（1秒以内を目指す）
func main() {
    // 1. 認証トークン取得（~100ms）
    token, err := auth.GetToken()

    // 2. リポジトリ検出（~50ms）
    repo, err := repo.Detect()

    // 3. TUI起動（即座）
    app := app.New(token, repo)

    // 4. データ取得は起動後に非同期実行
    p := tea.NewProgram(app)
    p.Run()
}

// App.Init() でデータ取得開始
func (a *App) Init() tea.Cmd {
    return tea.Batch(
        a.fetchWorkflows(),  // 非同期
        a.spinner.Tick,
    )
}
```

### ログメモリ管理

```go
const MaxLogLines = 10000  // 最大保持行数

func (l *LogViewport) SetContent(content string) {
    lines := strings.Split(content, "\n")
    if len(lines) > MaxLogLines {
        // 古いログを切り捨て
        lines = lines[len(lines)-MaxLogLines:]
    }
    l.viewport.SetContent(strings.Join(lines, "\n"))
}
```

### API制限対策

```go
// GitHub API: 5000 req/hour
// ポーリング間隔2秒 = 1800 req/hour（余裕あり）

// 実行中のRunのみポーリング
func (a *App) shouldPoll() bool {
    if run, ok := a.runs.Selected(); ok {
        return run.IsRunning()
    }
    return false
}
```

---

## 次のステップ

1. プロジェクト初期化
2. `auth/token_test.go` を書く（最初のテスト）
3. TDDサイクル開始

---

## 参考リソース

### lazydocker (jesseduffield/lazydocker)

以下のパターンを参考にした:

| パターン | lazydocker | lazyactions |
|---------|-----------|-------------|
| **TickerTask** | `NewTickerTask(200ms)` でDocker状態をポーリング | `TickerTask(2s)` でログをポーリング |
| **FilteredList[T]** | コンテナ/イメージのフィルタリング | Workflow/Run/Jobのフィルタリング |
| **Autoscroll** | ログ表示で最下部自動追従 | ログビューポートで自動追従 |
| **SideListPanel[T]** | 左サイドパネルからメイン表示へ | 3ペイン構成で同様の連携 |
| **Context cancellation** | ゴルーチンの適切な終了 | ポーリング停止時に使用 |

### gama (termkit/gama)

GitHub Actions TUIの先行実装として参考。ただし以下を改善:

| 項目 | gama | lazyactions |
|-----|------|-------------|
| ログ表示 | なし（ブラウザ遷移） | リアルタイムストリーミング |
| UIスタイル | タブベース | lazygit風3ペイン |
| Jobs API | 未使用 | ステップ詳細表示 |

### BubbleTea/Lipgloss (charmbracelet)

TUIフレームワークとして使用:
- `bubbletea`: tea.Model, tea.Cmd, tea.Msg パターン
- `lipgloss`: スタイリング
- `bubbles/viewport`: ログ表示用スクロール可能ビュー
- `bubbles/textinput`: フィルター入力用

---

## 実装TODOリスト

### Phase 1: MVP（3週間）

#### Week 1: 基盤構築

- [ ] **Day 1: プロジェクト初期化**
  - [ ] `go mod init github.com/nnnkkk7/lazyactions`
  - [ ] Makefile作成（test, build, lint, cover）

- [ ] **Day 2: auth/ -認証**
  - [ ] `token_test.go`: gh CLI優先、ENV fallbackのテスト
  - [ ] `token.go`: GetToken(), SecureToken実装
  - [ ] カバレッジ80%達成

- [ ] **Day 3: repo/ -リポジトリ検出**
  - [ ] `detect_test.go`: SSH/HTTPS URL解析テスト
  - [ ] `detect.go`: Detect(), parseGitHubURL()実装
  - [ ] `validate.go`: 入力バリデーション
  - [ ] カバレッジ85%達成

- [ ] **Day 4-5: github/ -APIクライアント**
  - [ ] `interface.go`: Client インターフェース定義
  - [ ] `types.go`: Workflow, Run, Job, Step型
  - [ ] `errors.go`: AppError, WrapAPIError
  - [ ] `client_test.go`: httptestでモックサーバー
  - [ ] `client.go`: ListWorkflows, ListRuns実装
  - [ ] `mock.go`: MockClient（呼び出し追跡付き）
  - [ ] カバレッジ75%達成

#### Week 2: コア機能

- [ ] **Day 1: github/ -ログ関連API**
  - [ ] `client.go`: ListJobs, GetJobLogs実装
  - [ ] `sanitize.go`: ログサニタイズ
  - [ ] `sanitize_test.go`: シークレット検出テスト

- [ ] **Day 2-3: app/ -3ペインレイアウト**
  - [ ] `styles.go`: Lipglossスタイル定義
  - [ ] `keymap.go`: キーバインド定義
  - [ ] `list.go`: FilteredList[T]
  - [ ] `app.go`: 基本3ペインUI
  - [ ] `app_test.go`: ペイン移動、選択テスト

- [ ] **Day 4-5: app/ -ログストリーミング（★最重要）**
  - [ ] `ticker.go`: TickerTask
  - [ ] `logview.go`: LogViewport（Autoscroll付き）
  - [ ] `polling.go`: AdaptivePoller
  - [ ] `messages.go`: BubbleTeaメッセージ型
  - [ ] `commands.go`: 不変値キャプチャ付きコマンド
  - [ ] ポーリング開始/停止テスト

#### Week 3: 機能完成

- [ ] **Day 1: アクション機能**
  - [ ] キャンセル（確認ダイアログ付き）
  - [ ] 再実行（全体/失敗のみ）
  - [ ] workflow_dispatch（inputs対応）

- [ ] **Day 2: フィルター・コピー**
  - [ ] `/`キーでフィルターモード
  - [ ] `y`キーでクリップボードコピー
  - [ ] 失敗ステップ自動ジャンプ

- [ ] **Day 3: UI仕上げ**
  - [ ] `?`ヘルプポップアップ
  - [ ] コンテキスト依存ステータスバー
  - [ ] レート制限表示
  - [ ] ローディングスピナー
  - [ ] フラッシュメッセージ

- [ ] **Day 4-5: リリース準備**
  - [ ] 全体カバレッジ70%確認
  - [ ] goreleaser設定
  - [ ] Homebrew Formula
  - [ ] README.md作成

### Phase 2: 運用機能（将来）

- [ ] マルチリポジトリ対応
- [ ] アーティファクト管理
- [ ] Runner状態監視
- [ ] 設定ファイル対応
- [ ] ワークフローYAML表示

---

## 全体構造図

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           lazyactions アーキテクチャ                         │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                                cmd/lazyactions                               │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ main.go                                                               │  │
│  │  1. auth.GetToken() → SecureToken                                     │  │
│  │  2. repo.Detect() → Repository                                        │  │
│  │  3. app.New(client, repo) → App                                       │  │
│  │  4. tea.NewProgram(app).Run()                                         │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                                     app/                                     │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                              App (tea.Model)                            ││
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐                       ││
│  │  │ workflows   │ │ runs        │ │ jobs        │  FilteredList[T]      ││
│  │  └─────────────┘ └─────────────┘ └─────────────┘                       ││
│  │  ┌─────────────────────────────────────────────┐                       ││
│  │  │ logView (LogViewport)                       │  Autoscroll対応       ││
│  │  └─────────────────────────────────────────────┘                       ││
│  │  ┌─────────────────────────────────────────────┐                       ││
│  │  │ logPoller (TickerTask) + AdaptivePoller     │  2秒ポーリング        ││
│  │  └─────────────────────────────────────────────┘                       ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐               │
│  │ keymap.go  │ │ styles.go  │ │messages.go │ │commands.go │               │
│  └────────────┘ └────────────┘ └────────────┘ └────────────┘               │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                      ┌────────────────┼────────────────┐
                      ▼                ▼                ▼
┌──────────────────────────┐ ┌──────────────────┐ ┌──────────────────────────┐
│        github/           │ │      auth/       │ │         repo/            │
│ ┌──────────────────────┐ │ │ ┌──────────────┐ │ │ ┌──────────────────────┐ │
│ │ Client (interface)   │ │ │ │ SecureToken  │ │ │ │ Detect()             │ │
│ │  - ListWorkflows()   │ │ │ │ GetToken()   │ │ │ │ parseGitHubURL()     │ │
│ │  - ListRuns()        │ │ │ │  1. gh CLI   │ │ │ │  - SSH形式           │ │
│ │  - ListJobs()        │ │ │ │  2. ENV      │ │ │ │  - HTTPS形式         │ │
│ │  - GetJobLogs()      │ │ │ └──────────────┘ │ │ └──────────────────────┘ │
│ │  - CancelRun()       │ │ └──────────────────┘ │ ┌──────────────────────┐ │
│ │  - RerunWorkflow()   │ │                      │ │ ValidateOwner()      │ │
│ │  - TriggerWorkflow() │ │                      │ │ ValidateRepoName()   │ │
│ └──────────────────────┘ │                      │ └──────────────────────┘ │
│ ┌──────────────────────┐ │                      └──────────────────────────┘
│ │ types.go             │ │
│ │  - Workflow          │ │
│ │  - Run               │ │
│ │  - Job, Step         │ │
│ └──────────────────────┘ │
│ ┌──────────────────────┐ │
│ │ errors.go            │ │
│ │  - AppError          │ │
│ │  - WrapAPIError()    │ │
│ └──────────────────────┘ │
│ ┌──────────────────────┐ │
│ │ sanitize.go          │ │
│ │  - SanitizeLogs()    │ │
│ └──────────────────────┘ │
│ ┌──────────────────────┐ │
│ │ mock.go              │ │
│ │  - MockClient        │ │
│ │  - 呼び出し追跡      │ │
│ └──────────────────────┘ │
└──────────────────────────┘
            │
            ▼
┌──────────────────────────┐
│    GitHub Actions API    │
│  go-github/v68           │
└──────────────────────────┘
```

### データフロー

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              BubbleTea MVU パターン                          │
└─────────────────────────────────────────────────────────────────────────────┘

User Input                    Model                         View
    │                           │                             │
    │  tea.KeyMsg               │                             │
    ├──────────────────────────►│                             │
    │                           │  Update()                   │
    │                           │  ├─ キー処理                │
    │                           │  ├─ 状態更新                │
    │                           │  └─ tea.Cmd 発行            │
    │                           │                             │
    │                           │  tea.Cmd                    │
    │                           ├─────────────────────────────┤
    │                           │  (非同期API呼び出し)        │
    │                           │                             │
    │                           │  tea.Msg (結果)             │
    │                           │◄────────────────────────────┤
    │                           │                             │
    │                           │  Update()                   │
    │                           │  └─ 結果を状態に反映        │
    │                           │                             │
    │                           │                             │
    │                           │────────────────────────────►│
    │                           │            View()           │
    │                           │            └─ UI描画        │
    │◄─────────────────────────────────────────────────────────│
    │                    画面更新                              │
```

### メッセージフロー例（ログストリーミング）

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         ログストリーミング フロー                            │
└─────────────────────────────────────────────────────────────────────────────┘

1. Run選択
   ┌──────────┐      ┌──────────┐      ┌──────────┐
   │ KeyEnter │ ──► │ Update() │ ──► │fetchJobs()│
   └──────────┘      └──────────┘      └──────────┘
                                              │
                                              ▼
2. Jobs取得完了                        ┌──────────────┐
   ┌──────────────┐      ┌──────────┐ │ ListJobs API │
   │JobsLoadedMsg │ ◄── │ 非同期   │◄┤              │
   └──────────────┘      └──────────┘ └──────────────┘
          │
          ▼
3. ログ取得開始
   ┌──────────┐      ┌────────────────┐
   │ Update() │ ──► │startAdaptive() │
   └──────────┘      └────────────────┘
                            │
                            ▼
4. ポーリングループ   ┌─────────────────┐
                      │ TickerTask      │
                      │ ┌─────────────┐ │
                      │ │ 2秒待機     │ │
                      │ └─────────────┘ │
                      │       │         │
                      │       ▼         │
                      │ ┌─────────────┐ │
                      │ │GetJobLogs() │ │
                      │ └─────────────┘ │
                      │       │         │
                      │       ▼         │
                      │ ┌─────────────┐ │
                      │ │LogsLoadedMsg│──────► Update()
                      │ └─────────────┘ │         │
                      └─────────────────┘         ▼
                                           ┌──────────────┐
5. ログ表示更新                            │logView更新   │
   ┌──────────────┐                        │Autoscroll    │
   │ View()       │◄───────────────────────│失敗ジャンプ  │
   └──────────────┘                        └──────────────┘
          │
          ▼
   ┌──────────────────────────────────────────────────────┐
   │ ╭─ Logs ───────────────────────────────────────────╮ │
   │ │  ✓ Checkout                                      │ │
   │ │  ● Run tests                          [running]  │ │
   │ │    === RUN   TestFoo                             │ │
   │ │    --- PASS: TestFoo                             │ │
   │ │    ▌                      ◄── Autoscroll位置     │ │
   │ ╰──────────────────────────────────────────────────╯ │
   └──────────────────────────────────────────────────────┘

6. Run完了時
   ┌────────────────┐      ┌──────────────┐
   │ Run.Status=    │ ──► │ ポーリング   │
   │ "completed"    │      │ 自動停止     │
   └────────────────┘      └──────────────┘
```

### ファイル依存関係

```
cmd/lazyactions/main.go
    │
    ├── auth/token.go
    │       └── SecureToken
    │
    ├── repo/detect.go
    │       ├── validate.go
    │       └── Repository
    │
    └── app/app.go
            ├── github/interface.go (Client)
            │       ├── client.go (実装)
            │       ├── types.go (Workflow, Run, Job, Step)
            │       ├── errors.go (AppError)
            │       ├── sanitize.go
            │       └── mock.go (テスト用)
            │
            ├── keymap.go
            ├── styles.go
            ├── messages.go
            ├── commands.go
            ├── list.go (FilteredList[T])
            ├── ticker.go (TickerTask)
            ├── logview.go (LogViewport)
            ├── polling.go (AdaptivePoller)
            ├── clipboard.go
            └── trigger.go (TriggerState)
```
