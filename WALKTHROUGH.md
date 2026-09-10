# Turbo IDE 제품군 개선 완료 보고서

대상 프로젝트:
- **Turbo Go (`tg`)**: `/Users/maro/Projects/tg`
- **Turbo Rust (`tr`)**: `/Users/maro/Projects/tr`
- **Turbo FORTRAN 77 (`tf77`)**: `/Users/maro/Projects/tf77`

---

## 1. 키보드 커서 시인성 개선 (Cursor Visibility Fix)

### 문제 배경
- 볼랜드 정통 에디터 배경색인 짙은 파란색(Borland Blue, `#0000A8`) 환경에서 터미널 기본 하드웨어 커서(특히 macOS Terminal.app, iTerm2, VS Code 통합 터미널 등)가 검은색 또는 1픽셀 두께의 얇은 선으로 렌더링되어 커서의 현재 위치 파악이 어려웠습니다.

### 2단계 하이브리드 커서 시인성 솔루션 적용
1. **터미널 하드웨어 커서 스타일 & 색상 제어**:
   - `tcell.Screen.SetCursorStyle(tcell.CursorStyleBlinkingBlock, tcell.ColorYellow)` 적용.
   - VT100 / xterm 확장 ANSI 이스케이프 시퀀스 출력:
     - `\x1b]12;#FFFF00\x07` (OSC 12: 커서 색상을 밝은 노란색 `#FFFF00`으로 변경)
     - `\x1b[1 q` (DECSCUSR 1: 블랭킹 블록 커서)
   - 앱 종료(`app.Stop()`) 시 `\x1b]112\x07\x1b[0 q`를 전송하여 사용자 터미널의 원래 커서 색상과 모양으로 안전하게 복원.

2. **소프트웨어 에디터 버퍼 커서 하이라이트 (`editor.Draw`)**:
   - 하드웨어 이스케이프 시퀀스를 지원하지 않는 환경에서도 100% 확실한 시인성을 보장하도록 에디터 버퍼의 커서 셀 자체를 **밝은 노란색 배경(`ColorEditorCursor` / `#FFFF00`) + 검은색 텍스트(`tcell.ColorBlack`)**로 직접 렌더링.
   - 탭 문자(`\t`) 확장 및 멀티바이트 룬 폭(`runewidth.RuneWidth`)을 정밀 계산하여 실제 화면 좌표(`actualCursorScreenX`)에 커서 셀 배치.
   - 빈 줄이나 라인 끝(EOL) 너머 공백 위치에서도 노란색 블록 커서 셀 표시.
   - 거터(Gutter) 라인 번호 영역에서 현재 커서가 위치한 행을 **밝은 노란색 굵은 글씨**와 **`▸` 마커**(`▸ 12 `)로 강조 표시.

3. **다이얼로그 및 모달 포커스 관리**:
   - 텍스트 입력 다이얼로그(`FindDialog`, `SaveFileDialog`, `GotoLineDialog`): 텍스트 입력 필드에 노란색 커서 블록 표시 및 `screen.ShowCursor(curX, curY)` 연동.
   - 비입력/조회 다이얼로그(`AboutDialog`, `CompileDialog`, `ErrorListDialog`, `OpenFileDialog`, `SearchResultsDialog`): `IsVisible() bool` 인터페이스 구현 및 `screen.HideCursor()` 호출로 불필요한 커서 숨김 처리.
   - 메인 루프 렌더링 시 모달 다이얼로그나 메뉴바가 활성화된 경우 에디터 커서를 숨겨 포커스 혼선 방지.

---

## 2. 멀티 파일 프로젝트 심볼 검색 및 정의 이동 (Go to Definition)

### 1. 커서 위치 식별자 단어 자동 감지 (`GetWordUnderCursor`)
- 에디터 버퍼의 현재 커서(`CursorX`, `CursorY`) 위치에서 식별자 문자(`[a-zA-Z0-9_]`)를 좌우로 탐색하여 단어를 자동 추출.
- `Ctrl+F` (Find), `Alt+F3` (Find in Project), `F12` (Go to definition) 실행 시 검색 대상 단어로 자동 채움.

### 2. 정의로 이동 (`F12` / Go to Definition)
- 에디터에서 함수명, 타입명, 서브루틴명 위에 커서를 두고 **`F12`**를 누르면, 프로젝트 내 모든 소스 파일을 스캔하여 정의된 위치로 즉시 이동.
- 정의가 다른 파일에 존재하는 경우, 자동으로 해당 파일을 에디터에 로드한 후 정확한 줄 번호(`Line`)와 컬럼(`Column`)으로 커서 이동.

### 3. 프로젝트 전체 검색 (`Alt+F3` / Find in Project)
- `Search ➔ Find in project...` (`Alt+F3`) 실행 시 대소문자 구분 옵션을 지원하며 전체 소스 파일 검색.
- 검색 결과 창(`SearchResultsDialog`)에서 화살표 키로 탐색 및 `Enter` 입력으로 해당 위치 점프.

---

## 프로젝트별 수정 파일 내역

| 프로젝트 | 수정 파일 |
| :--- | :--- |
| **`tg` (Turbo Go)** | `internal/ui/app.go`<br>`internal/ui/editor.go`<br>`internal/ui/dialogs/find.go`<br>`internal/ui/dialogs/savefile.go`<br>`internal/ui/dialogs/gotoline.go`<br>`internal/ui/dialogs/about.go`<br>`internal/ui/dialogs/compile.go`<br>`internal/ui/dialogs/errorlist.go`<br>`internal/ui/dialogs/openfile.go`<br>`internal/ui/dialogs/search_results.go`<br>`cmd/tg/main.go` |
| **`tr` (Turbo Rust)** | `internal/ui/app.go`<br>`internal/ui/editor.go`<br>`internal/ui/dialogs/find.go`<br>`internal/ui/dialogs/savefile.go`<br>`internal/ui/dialogs/gotoline.go`<br>`internal/ui/dialogs/about.go`<br>`internal/ui/dialogs/compile.go`<br>`internal/ui/dialogs/errorlist.go`<br>`internal/ui/dialogs/openfile.go`<br>`internal/ui/dialogs/search_results.go`<br>`cmd/tr/main.go` |
| **`tf77` (Turbo FORTRAN 77)** | `internal/ui/app.go`<br>`internal/ui/editor.go`<br>`internal/ui/dialogs/find.go`<br>`internal/ui/dialogs/savefile.go`<br>`internal/ui/dialogs/gotoline.go`<br>`internal/ui/dialogs/about.go`<br>`internal/ui/dialogs/compile.go`<br>`internal/ui/dialogs/errorlist.go`<br>`internal/ui/dialogs/openfile.go`<br>`internal/ui/dialogs/search_results.go`<br>`cmd/tf77/main.go` |

---

## 검증 및 빌드 결과

3개 프로젝트 모두 단위 테스트 통과 및 바이너리 빌드를 완료하였습니다:

```bash
# tg (Turbo Go)
$ cd /Users/maro/Projects/tg && go test ./... && go build -o bin/tg ./cmd/tg
ok  	tg/internal/compiler	(cached)
ok  	tg/internal/debugger	(cached)
ok  	tg/internal/syntax	(cached)
ok  	tg/internal/ui	(cached)

# tr (Turbo Rust)
$ cd /Users/maro/Projects/tr && go test ./... && go build -o bin/tr ./cmd/tr
ok  	tr/internal/compiler	(cached)
ok  	tr/internal/debugger	(cached)
ok  	tr/internal/syntax	(cached)
ok  	tr/internal/ui	(cached)

# tf77 (Turbo FORTRAN 77)
$ cd /Users/maro/Projects/tf77 && go test ./... && go build -o bin/tf77 ./cmd/tf77
ok  	tf77/internal/compiler	(cached)
ok  	tf77/internal/debugger	(cached)
ok  	tf77/internal/syntax	(cached)
ok  	tf77/internal/ui	(cached)
```
