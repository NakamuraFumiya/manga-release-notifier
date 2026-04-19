# OOP 原則の採用方針

## このドキュメントの目的

本プロジェクト (`manga-release-notifier`) のドメイン層を設計・実装するにあたり、採用する OOP 原則と、それを Go でどう表現するかをまとめる。

本ドキュメントは **座学メモ** と **設計判断の記録** を兼ねる。各セクションは学習の進捗に合わせて段階的に埋めていく。最初から完璧に書こうとせず、「その原則を意識した実装をした時点で、該当セクションに具体例と学びを追記する」運用とする。

## なぜ OOP を意識するか

- Ruby キャリア開始時は OOP を十分に理解せず使っていた
- Go でも 4 年間 OOP を意識せず書いてきた
- 本プロジェクトを通じて OOP の原則を体系的に身につける
- Go に OOP を適用できることを実体験する(継承のない言語でも OOP の本質は表現できる)

詳細な動機と代替案の却下理由は `docs/manga-notifier-spec.md` の ADR-013 を参照。

## OOP の 4 本柱

OOP を構成する 4 つの基礎概念。どれも「ソフトウェアを変更しやすく保つ」ための道具。

### カプセル化 (Encapsulation)

**概要**: データと、そのデータを扱うロジックを一つの単位にまとめ、外部から触れる範囲を絞る。

**Go での表現**: 構造体のフィールドを小文字始まり(unexported)にして、公開メソッド経由でのみ操作できるようにする。

**学びメモ**: TODO(学習時に追記)

### 継承 (Inheritance)

**概要**: 既存のクラスから派生クラスを作り、振る舞いを引き継ぐ。OOP の古典的な道具だが、乱用すると脆い設計になる。

**Go での扱い**: Go には継承が **存在しない**。代替手段は「合成(has-a)」と「インターフェース」。本プロジェクトでは継承を使わず、合成で表現する。

**学びメモ**: TODO

### 多態性 (Polymorphism)

**概要**: 同じインターフェースで異なる実装を切り替えられる性質。OOP の柔軟性の核。

**Go での表現**: インターフェース(暗黙実装)で表現する。Ruby のダックタイピングとも近いが、Go は型チェックが働く分、安全。

**学びメモ**: TODO

### 抽象化 (Abstraction)

**概要**: 複雑な実装の詳細を隠し、利用者に必要な操作だけを見せる。

**Go での表現**: 小さいインターフェースを定義し、実装は別パッケージに閉じ込める(ドメイン層は infra 層を知らない、等)。

**学びメモ**: TODO

## SOLID 原則

クラス設計の実務的な 5 原則。Robert C. Martin が整理した。Go でも原則の本質は共通。

### S: Single Responsibility Principle (単一責任原則)

**概要**: 一つの型は一つの責務だけを持つ。変更理由が 2 つ以上ある型は分割する。

**学びメモ**: TODO

### O: Open/Closed Principle (開放閉鎖原則)

**概要**: 拡張には開き、変更には閉じる。新しい振る舞いを追加する時、既存コードを書き換えずに済む設計にする。

**学びメモ**: TODO

### L: Liskov Substitution Principle (リスコフの置換原則)

**概要**: 派生型は基底型と置き換え可能であるべき。Ruby の `class Penguin < Bird` で `fly` が壊れる例が典型的な違反。

**Go での位置づけ**: 継承がないので「派生型が基底型を壊す」問題は発生しないが、インターフェースの実装が契約を守っていない(期待される挙動を満たさない)ケースには適用できる。

**学びメモ**: TODO

### I: Interface Segregation Principle (インターフェース分離原則)

**概要**: 巨大なインターフェースを押し付けず、クライアントが必要なメソッドだけを持つ小さなインターフェースを提供する。

**Go との親和性**: 高い。Go コミュニティの「インターフェースは小さく」文化(例: `io.Reader`, `io.Writer`)そのもの。

**学びメモ**: TODO

### D: Dependency Inversion Principle (依存性逆転原則)

**概要**: 上位モジュール(ドメイン)は下位モジュール(infra)に依存すべきでなく、両者は抽象(インターフェース)に依存すべき。

**Go での表現**: ドメイン層がインターフェースを定義し、infra 層がそれを実装する。コンストラクタ経由でインターフェースを注入する(DI)。

**学びメモ**: TODO

## 「継承より合成 (Composition over Inheritance)」と Go

**概要**: Gang of Four (1994) で提唱された原則。継承は強い結合を生み、脆い基底クラス問題や LSP 違反を招きやすい。「is-a」より「has-a」を優先する。

**Go での位置づけ**: Go はそもそも継承を持たないため、この原則が言語レベルで強制されている。構造体の埋め込みとインターフェースによる合成で、継承の代替を行う。

**具体例**:

```go
// 継承的な発想 (Go には書けないが、仮想的にはこう書きたくなる)
// type Penguin extends Bird  ← Go にはない

// 合成による代替: 振る舞いを部品として持つ
type Movement interface {
    Move() string
}

type Bird struct {
    Movement Movement  // 部品として持つ (has-a)
}
```

**学びメモ**: TODO

## 本プロジェクトへの適用方針

Phase 1 以降の実装で下した設計判断と、それが依拠した OOP 原則を記録する。

### Phase 1

TODO

### Phase 2

TODO

### Phase 3

TODO

### Phase 4

TODO

## 参考資料

### OOP の原点

- **Alan Kay** — 「OOP」という用語の提唱者、Smalltalk の設計者
  - 論文: "The Early History of Smalltalk" (1993, *ACM SIGPLAN Notices* Vol. 28, No. 3)
    - DOI: `10.1145/155360.155364`
    - 要点: OOP は当初「メッセージパッシング」が本質であり、クラスや継承は副次的だったと語る一次資料
  - FAQ: "Dr. Alan Kay on the Meaning of 'Object-Oriented Programming'" (2003)
    - 有名な一文: *"I made up the term 'object-oriented', and I can tell you I did not have C++ in mind."*
- **Ole-Johan Dahl, Kristen Nygaard** — Simula-67 の設計者（OOP の言語的起源、クラスの概念の導入）
  - レポート: "SIMULA 67 COMMON BASE LANGUAGE" (1968, Norwegian Computing Center)

### デザインパターンと「継承より合成」

- **Erich Gamma, Richard Helm, Ralph Johnson, John Vlissides (Gang of Four)**
  - 書籍: 『**Design Patterns: Elements of Reusable Object-Oriented Software**』 (1994, Addison-Wesley)
  - 序文で "Favor object composition over class inheritance." を明示。本プロジェクトの方針の原典

### SOLID 原則

- **Robert C. Martin (Uncle Bob)** — SOLID 原則の整理・命名者（頭字語は Michael Feathers が命名）
  - 論文: "Design Principles and Design Patterns" (2000, Object Mentor)
    - SOLID 原則の原型となる 5 原則が提示された最初の文書
  - 書籍: 『**Clean Architecture: A Craftsman's Guide to Software Structure and Design**』 (2017, Prentice Hall)
  - 書籍: 『**Clean Code**』 (2008, Prentice Hall)
- 各原則の原典
  - **OCP (Open/Closed Principle)**: **Bertrand Meyer**『Object-Oriented Software Construction』 (1988 初版、1997 第2版、Prentice Hall)
  - **LSP (Liskov Substitution Principle)**: **Barbara Liskov** "Data Abstraction and Hierarchy" (1987, *ACM SIGPLAN Notices* Vol. 23, No. 5, OOPSLA keynote)
    - DOI: `10.1145/62138.62141`
  - **SRP / ISP / DIP**: 上記 Robert C. Martin の論文と書籍で定式化

### Go の OOP 観

- **Rob Pike** "Go at Google: Language Design in the Service of Software Engineering" (2012)
  - https://go.dev/talks/2012/splash.article
  - Go がなぜ継承を採用しなかったか、インターフェースをなぜ暗黙実装にしたかの設計判断が語られている
- **Effective Go** — Go 公式の慣用表現ガイド
  - https://go.dev/doc/effective_go
  - 特に "Interfaces" セクション: https://go.dev/doc/effective_go#interfaces

### ドメインモデリング / リファクタリング

- **Eric Evans**『**Domain-Driven Design: Tackling Complexity in the Heart of Software**』 (2003, Addison-Wesley)
- **Martin Fowler**『**Refactoring: Improving the Design of Existing Code**』 (1999 初版、2018 第2版、Addison-Wesley)
- **Martin Fowler**『**Patterns of Enterprise Application Architecture**』 (2002, Addison-Wesley)

### 読む順序の推奨

1. **Effective Go - Interfaces** (無料・短時間) — Go の前提を押さえる
2. **Rob Pike "Go at Google"** (無料・短時間) — Go の設計思想を知る
3. **Robert C. Martin "Design Principles and Design Patterns" (2000 論文)** (無料) — SOLID の全体像
4. **Robert C. Martin『Clean Architecture』** — 体系的な整理
5. **Gang of Four『Design Patterns』** — パターンの原典。全部読まず索引的に使う
6. **Barbara Liskov / Bertrand Meyer の原典** — 興味が出たら深掘り
