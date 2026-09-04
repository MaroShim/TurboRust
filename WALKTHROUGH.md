# Turbo Rust (`tr`) 프로젝트 구축 및 워크스루 🚀

볼랜드 Turbo Vision TUI 아키텍처와 감성을 완벽하게 계승하면서, **구현 언어는 Go로 구축하고 개발 대상 타깃 언어를 Rust로 전면 특화**한 레트로 TUI IDE **Turbo Rust (`tr`)** 구축을 완료하였습니다.

---

## 📸 프로젝트 핵심 아키텍처 및 구현 컴포넌트

```
c:/AntiGravity/TurboRust/
├── Cargo.toml                  # (참조/연동용)
├── go.mod                      # Go 모듈 정의 (tr, go 1.26+)
├── go.sum                      # 의존성 체크섬
├── README.md                   # Turbo Rust 공식 영문/국문 안내서
├── WALKTHROUGH.md              # 프로젝트 워크스루 문서
├── cmd/
│   └── tr/
│       └── main.go             # IDE 진입점, 이벤트 루프, 단축키 디스패처
├── examples/
│   ├── hello.rs                # 기본 출력 예제
│   └── fibonacci.rs            # 함수 및 반복문 예제
└── internal/
    ├── compiler/
    │   ├── runner.go           # rustc 및 cargo 빌드/실행/에러 파서
    │   └── runner_test.go      # 단위 테스트 및 실제 rustc 통합 테스트
    ├── debugger/
    │   ├── debugger.go         # 브레이크포인트, 실행 포인터, 변수 Watches 엔진
    │   └── debugger_test.go    # 디버거 세션 및 브레이크포인트 테스트
    ├── sound/
    │   └── sound.go            # 2.5인치 종이 콘 PC 스피커 물리 모델 사운드 합성기
    ├── syntax/
    │   ├── rust_highlighter.go # Rust 2021/2024 구문 강조기
    │   └── rust_highlighter_test.go
    └── ui/
        ├── app.go              # UI 애플리케이션 상태 컨트롤러
        ├── clipboard.go        # OS 및 내부 클립보드 브리지
        ├── editor.go           # 시그니처 터보 블루 에디터 버퍼
        ├── editor_test.go      # 에디터 조작 및 블록 선택 테스트
        ├── menubar.go          # 상단 볼랜드 풀다운 메뉴바
        ├── statusbar.go        # 하단 단축키 바
        ├── theme.go            # 볼랜드 색상 팔레트 및 이중선 박스 드로잉 기호
        ├── userscreen.go       # Alt+F5 전체화면 DOS 콘솔 뷰어
        ├── watchwindow.go      # Alt+W 변수 감시 윈도우
        ├── window.go           # 프레임, 다이얼로그 박스, 그림자 드로잉
        └── dialogs/
            ├── about.go        # Turbo Rust 정보 대화상자
            ├── compile.go      # "Compiling..." 통계 팝업 모달
            ├── errorlist.go    # 컴파일 에러 목록 및 에디터 즉시 점프
            ├── find.go         # 문자열 검색 대화상자
            ├── gotoline.go     # 특정 줄 번호 이동
            ├── openfile.go     # .rs 필터링 파일 브라우저
            └── savefile.go     # 파일 저장 대화상자
```

---

## 🛠️ 주요 기능 상세

### 1. Classic Borland Turbo Vision UI
- **시그니처 블루 캔버스 (`#0000A8`)**와 정교한 이중선 프레임 (`╔═╗`, `║ ║`, `╚═╝`)
- 입체 텍스트 드롭 섀도우(오른쪽 2칸, 아래 1행 그림자 효과)
- 윈도우 헤더: `[■] 1 NONAME00.RS [▲]` (닫기/최대화 버튼 및 윈도우 번호)
- 우측 수직 스크롤 트랙(`░`), 엄지 블록(`█`), 상/하 화살표(`▲`, `▼`)
- 상단 풀다운 메뉴바 (`File`, `Edit`, `Search`, `Run`, `Compile`, `Debug`, `Options`, `Window`, `Help`)
- 하단 볼랜드 핫키 바 (`F1 Help`, `F2 Save`, `F3 Open`, `Alt+F9 Compile`, `F9 Make`, `Ctrl+F9 Run`, `Alt+F5 User`, `F10 Menu`)

### 2. Rust 전용 구문 강조기 (Syntax Highlighter)
- **Rust 키워드**: `fn`, `let`, `mut`, `match`, `if`, `else`, `loop`, `while`, `for`, `in`, `return`, `struct`, `enum`, `impl`, `trait`, `pub`, `use`, `mod`, `crate`, `type`, `const`, `static`, `where`, `async`, `await` 등
- **기본 및 표준 타입**: `i8`~`i128`, `u8`~`u128`, `isize`, `usize`, `f32`, `f64`, `bool`, `char`, `str`, `String`, `Option`, `Result`, `Vec`, `Box`, `Rc`, `Arc` 등
- **매크로**: `println!`, `eprintln!`, `format!`, `vec!`, `panic!`, `assert!`, `todo!` 등 `!` 접미사
- **라이프타임**: `'a`, `'static`
- **리터럴**: Raw string (`r#"..."#`), 일반 문자열, 바이트 문자열, 진법별 숫자(16진수, 2진수, 부동소수점)
- **주석 및 속성**: 라인 주석(`//`), 멀티라인 블록 주석(`/* */`), 속성(`#[...]`)

### 3. 컴파일러 & 에러 진단 파싱
- **단일 파일**: `rustc --error-format=short -g -o <bin> <file.rs>` 초고속 컴파일
- **Cargo 프로젝트 자동 감지**: 상위 경로에 `Cargo.toml`이 있으면 `cargo build` 연동
- **에러 파싱 & 즉시 점프**: 컴파일 실패 시 파일명, 줄 번호, 열 번호, 에러 메시지를 파싱하여 에러 목록 팝업을 띄우고, 선택 시 **에디터 해당 줄과 컬럼으로 커서 즉시 점프**
- **볼랜드 "Compiling..." 모달**: Main file, rustc -> binary, Total lines, Errors, Warnings, Elapsed Time 표시

### 4. Alt+F5 User Screen (DOS 콘솔 실행 화면)
- `Ctrl+F9` 실행 시 결과 화면을 별도의 전체화면 콘솔 버퍼에 캡처하여 표시
- `[ Turbo Rust User Screen (Alt+F5) - Press any key to return to IDE ]` 배너
- 아무 키나 누르면 즉시 이전 에디터 상태로 복귀

### 5. 인터랙티브 디버거 & Watches 창
- `F4`: 브레이크포인트(`●`, 라인 전체 **솔리드 레드 바**)
- `F5` / `F7` / `F8`: Debug / Continue / Trace Into / Step Over (현재 실행 라인 **솔리드 옐로우 바 `►`**)
- Watches 창: 하단 Watches 윈도우 (Debug 메뉴로 토글, 변수명, 타입, 값 실시간 감시)

### 6. 볼랜드 레트로 PC 스피커 사운드 FX
- 2.5인치 종이 콘 PC 스피커 물리 모델 (Square wave + IIR 저역 통과 필터 + Attack/Decay 엔벨로프)
- 컴파일 성공 시: 경쾌한 2단 상승 비프 (740Hz ➔ 1108Hz)
- 컴파일 실패 시: 묵직한 저음 버즈 (196Hz)
- 브레이크포인트 적중 시: 아날로그 피에조 클릭 사운드 (880Hz)

---

## ⌨️ 주요 단축키 요약

| 단축키 | 기능 | 설명 |
|---|---|---|
| **Alt + F** | **File Menu** | **File 메뉴 즉시 열기** |
| **Alt + E** | **Edit Menu** | Edit 메뉴 즉시 열기 |
| **Alt + S** | **Search Menu** | Search 메뉴 즉시 열기 |
| **Alt + R** | **Run Menu** | Run 메뉴 즉시 열기 |
| **Alt + C** | **Compile Menu** | Compile 메뉴 즉시 열기 |
| **Alt + D** | **Debug Menu** | Debug 메뉴 즉시 열기 |
| **Alt + O** | **Options Menu** | Options 메뉴 즉시 열기 |
| **Alt + W** | **Window Menu** | Window 메뉴 즉시 열기 |
| **Alt + H** | **Help Menu** | Help 메뉴 즉시 열기 |
| **F10** | **Menu Bar** | 상단 메뉴바 전체 포커스 및 File 메뉴 열기 |
| **F1** | Help / About | Turbo Rust 정보 팝업 |
| **F2** | Save | 현재 버퍼 저장 / 다른 이름으로 저장 |
| **F3** | Open | 파일 브라우저 (.rs 필터링) |
| **F4** | **Breakpoint** | 브레이크포인트 설정/해제 (레드 바 `●`) |
| **F5** | **Debug / Continue** | 디버깅 시작 / 다음 브레이크포인트까지 계속 실행 |
| **F7** | **Trace Into** | 한 줄씩 실행 (함수/명령 진입) |
| **F8** | **Step Over** | 한 줄씩 실행 (함수 건너뛰기) |
| **Alt + F9** | **Compile** | "Compiling..." 모달과 함께 빌드 |
| **F9** | Make | 빌드 실행 |
| **Ctrl + F9** | **Run** | 빌드 후 실행 및 User Screen 표시 |
| **Alt + F5** | **User Screen** | 프로그램 실행 결과 화면 토글 |
| **Alt + L** | Line Numbers | 줄 번호 거터 On/Off (`F6`) |
| **Alt + G** | Go to Line | 특정 줄 번호로 이동 (`Ctrl+G`) |
| **Ctrl + F** | Find | 문자열 검색 |
| **Ctrl + L** | Search Again | 다음 찾기 |
| **Alt + X** | Exit | Turbo Rust 종료 |

---

## 🧪 검증 결과

1. **단위 테스트**:
   - `go test ./...` ➔ 모든 테스트 패키지(`internal/syntax`, `internal/compiler`, `internal/debugger`, `internal/ui`) 100% PASS
2. **실제 rustc 통합 컴파일 & 실행 테스트**:
   - `internal/compiler/runner_test.go`에서 `rustc`로 `examples/hello.rs`를 컴파일하고 바이너리 실행 결과 출력 캡처 확인 (PASS, 2.05s)
3. **바이너리 빌드**:
   - `go build -o bin/tr.exe ./cmd/tr` ➔ 오류 없이 `bin/tr.exe` 성공적으로 생성

---

## 🚀 실행 가이드

```bash
# Turbo Rust 디렉터리로 이동
cd c:\AntiGravity\TurboRust

# 1. Turbo Rust 기본 실행
.\bin\tr.exe

# 2. 예제 코드 열기
.\bin\tr.exe examples\hello.rs
.\bin\tr.exe examples\fibonacci.rs
```
