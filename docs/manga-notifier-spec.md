# 漫画更新通知bot プロジェクト仕様書

## このドキュメントの目的

個人学習プロジェクトとして、好きなWeb漫画の更新を検知してDiscordに通知するbotを作る。AWSサーバーレス(Lambda + SQS + EventBridge + DynamoDB)で実プロダクトを本番稼働させつつ、同じconsumerをローカルのkind(k8s)でも動かせるようにして、「同じ問題をサーバーレスとk8sでどう解くか」を体で理解することがゴール。

業務で触っている history_job パターン(producer / queue / consumer の非同期アーキテクチャ)を個人プロジェクトで再現することで、学びを加速させる。

## このプロジェクトの大原則: 自分で手を動かす

**このプロジェクトの目的は「動くbotを作ること」ではなく「作る過程で学ぶこと」**。Claude Codeに全部書かせて完成させても学習にはならない。以下のルールを守ること。

### Claude Code への依頼ルール

**自分で書く(Claude Codeに書かせない)もの**:
- Go のコード全て(handler, domain, usecase, infra)
- Terraform のリソース定義(main.tf の中身)
- k8s manifest(Deployment, ScaledObject など)
- Dockerfile

**Claude Code に頼んでよいもの**:
- エラーメッセージの解読と原因の仮説出し
- 設計の壁打ち(「この場合AとBどっちがいい?」)
- 概念の説明(「KEDAのScaledObjectのpollingIntervalって何?」)
- 詰まった時のヒント(コードを書いてもらうのではなく、考え方を教えてもらう)
- コードレビュー(自分が書いたものに対するフィードバック)
- ドキュメントの雛形(README、振り返りメモのテンプレート等)

**グレーゾーン(最初は自分で、詰まったら相談)**:
- ボイラープレート的なコード(go.mod の初期化、LambdaのハンドラーのAWS SDK部分など)
- 最初は写経でもいいので自分でタイプする。どうしても分からなければ「こういう処理がしたいが、AWS SDK v2 でどう書くのがイディオムか?」と聞く

### 手を動かすための具体的なプラクティス

- **Phase ごとに "写経OKライン" を決める**: 例えば Phase 1 は「AWS Lambda Go SDK の使い方を初めて触るので公式ドキュメントから写経OK」、Phase 3 は「Phase 1 の経験があるので自力で書く」など、段階的に自力度を上げる
- **詰まったら15分ルール**: 15分考えて分からなければ Claude に聞く。ただし聞くのは「答え」ではなく「ヒント」
- **コピペ禁止の宣言**: Claude Code が出したコードは、写経する形で自分でタイプし直す。手が覚える効果は侮れない
- **振り返りを書く**: 各フェーズ完了時に `docs/phases/phaseN.md` に「詰まったところ」「新しく学んだこと」「次に活かしたいこと」を書く。これが Zenn 記事の下書きにもなる
- **AIに聞く前に公式ドキュメントを読む**: AWS のドキュメント、Terraform の provider ドキュメント、Go のパッケージドキュメントを先に読む癖をつける

### なぜこの原則が重要か

- ふみやさんの学習ロードマップ(Step 4: インフラ基礎)の先行投資としてこのプロジェクトがある。コードを書かなければインフラの「勘所」は身につかない
- 業務で感じている「インフラは読めるけど設計判断の why が分からない」というギャップは、自分で小さく作って初めて埋まる
- Claude Code を「頼れる同僚」として使うのと「代わりに全部やってくれる人」として使うのは全く違う。前者の使い方を身につけること自体が、現代のエンジニアリング学習の一部
- 後から記事化する時、「自分で書いたコード」でないと言語化できない部分が必ずある

## 学習ゴール

- AWS Lambda / SQS / EventBridge Scheduler / DynamoDB を Terraform で構築できるようになる
- Lambda の Event Source Mapping による SQS トリガーの挙動を理解する
- ローカル開発環境として LocalStack を使い、AWS料金ゼロで反復開発できる状態を作る
- kind + KEDA で SQS をトリガーに scale する k8s consumer を動かし、業務の KEDA 運用と地続きの理解を得る
- Go(業務と同じスタック)で producer / consumer を実装し、ドメイン設計を Clean Architecture で整理する
- IaC (Terraform) のモジュール分割と、環境ごとの provider 切り替え(本番AWS / LocalStack)を経験する
- OOP の原則(4本柱、SOLID、継承より合成)を意識したドメイン設計・実装で「Go における OOP」を体得する。詳細は `docs/oop-principles.md` を参照

## 最終的な構成イメージ

### 本番(AWS)

```
EventBridge Scheduler (1時間おき cron)
        ↓
   Lambda: fetcher (RSS取得 + 差分検知)
        ↓ (新着があれば publish)
   DynamoDB (最新話の状態保存)
        ↓
      SQS: manga-updates queue
        ↓ (Event Source Mapping)
   Lambda: notifier (Discord webhook)
        ↓
    Discord
```

### ローカル開発(LocalStack)

本番と同じterraformコードを、LocalStackに向けてapplyする。エンドポイントをlocalhost:4566に差し替えるだけで、同じLambda / SQS / DynamoDBが手元で動く。料金ゼロ。AWSリソースの反復開発はネット不要だが、RSS取得とDiscord通知の実動作確認にはネット接続が必要。

### ローカルk8s実験(kind + KEDA)

同じGoバイナリをDockerイメージ化して、kind上のPodとして動かす。KEDAのScaledObjectでSQS(LocalStackのSQS)のキュー深度をトリガーに scale-to-zero / scale-out する挙動を観察する。notifier Lambdaとk8s Podは排他ではなく、「同じメッセージをどちらでも処理できる2実装」として比較対象にする。

## 使用技術

- **言語**: Go 1.22+
- **IaC**: Terraform
- **AWSサービス**: Lambda, SQS, EventBridge Scheduler, DynamoDB, IAM, CloudWatch Logs
- **ローカルAWS**: LocalStack (Community Edition, 無料)
- **ローカルk8s**: kind, KEDA
- **RSSパース**: `github.com/mmcdole/gofeed`
- **AWS SDK**: `aws-sdk-go-v2`
- **通知先**: Discord webhook

## アーキテクチャ決定記録 (ADR)

各設計判断とその理由をここに残す。後から「なんでこうしたんだっけ」と迷わないために。

### ADR-001: なぜサーバーレス(Lambda)を中心に据えるか

**決定**: Producer/Consumerともに本番はLambdaで動かす。

**理由**:
- 個人プロジェクトで常時稼働させても月$1〜5程度に収まり、EKSの$73/月と比べて圧倒的に安い
- 「使っていない時は料金ゼロ」の設計体験はk8sでは得にくい
- 業務で触っていないサーバーレス特有の罠(cold start、実行時間制限、IAMロールの粒度)を学べる
- EventBridge Scheduler + Lambda + SQS は AWS サーバーレスの最も一般的な組み合わせで、実務でも頻出

**却下した代替案**:
- ECS Fargate: KEDA相当の学習ができない、k8s manifestを書く練習にならない
- EC2常時稼働: 学習ROIが低い、運用負荷が高い

### ADR-002: なぜ fetcher と notifier の2つのLambdaに分けるか

**決定**: RSS取得を行うLambdaと、Discord通知を行うLambdaを分離し、間にSQSを挟む。

**理由**:
- **責務分離**: fetcherはRSS取得と差分検知だけ、notifierは通知だけ。それぞれ単体でテストしやすい
- **リトライ独立性**: Discord通知だけ失敗した場合、RSS取得からやり直す必要がない。SQSのvisibility timeoutとDLQで通知側だけリトライできる
- **並行度制御**: notifierはDiscordのレート制限を考慮して同時実行数を絞りたいが、fetcherには関係ない。Lambda単位で reserved concurrency を設定できる
- **業務の history_job パターンの再現**: producer(fetcher)→queue(SQS)→consumer(notifier)という非同期アーキテクチャは、業務で使っている構成と同じ形。個人プロジェクトで追体験することで理解が深まる
- **将来の拡張性**: 通知先がSlackやメールに増えても、notifier Lambdaを追加するだけで fetcher は触らずに済む

**却下した代替案**:
- 1つのLambdaで完結: シンプルだが、学習目的(非同期パターンの体得)から外れる。実用性だけなら正解かもしれないが、今回のゴールはアーキテクチャ学習

### ADR-003: なぜ SQS → Lambda の間に EventBridge を挟まないか

**決定**: SQSからLambdaへは Event Source Mapping で直接繋ぐ。EventBridge Busは使わない。

**理由**:
- Lambda には SQS トリガー(Event Source Mapping)機能があり、SQSにメッセージが入ると自動でポーリングして起動してくれる
- EventBridge Busを挟む意味は「1つのイベントを複数の宛先に fan-out したい」時だが、現時点では宛先が1つ(Discord)なので不要
- シンプルな構成から始めて、必要になったら追加する原則
- EventBridge Schedulerは別の用途(fetcher の cron 起動)で使うので、「EventBridgeの2つの顔(Scheduler と Bus)」の違いはこのプロジェクトで学べる

**却下した代替案**:
- EventBridge Bus を間に挟む: 現時点では過剰設計。将来「Slackにも通知したい」となったらここに Bus を追加する拡張余地として覚えておく

### ADR-004: なぜ差分検知の状態を DynamoDB に持つか

**決定**: 各作品の「前回取得時の最新話」をDynamoDBに保存し、fetcherが比較する。

**理由**:
- Lambda はステートレスなので、実行間の状態をどこかに持つ必要がある
- DynamoDB は無料枠(25GBストレージ、25WCU/25RCU)が大きく、個人プロジェクトなら実質無料
- Key-Valueアクセスだけで済む(PK: 作品ID、値: 最新話情報)ので RDB は不要
- 業務では MySQL を使っているので、DynamoDB という別パラダイムを触る学習機会になる
- NoSQLの「単一テーブル設計」や「アクセスパターン駆動設計」を体験する入り口として最適

**却下した代替案**:
- RDS (Aurora Serverless v2): 無料枠が弱く、個人プロジェクトでは割高
- S3にJSONファイル: 可能だが、同時書き込みの整合性が弱い。学習教材としても薄い

### ADR-005: なぜ RSS提供サイトのみに絞るか

**決定**: 初期実装ではRSSを提供しているWeb漫画ポータル(少年ジャンプ+など)のみを対象にし、スクレイピングは行わない。

**理由**:
- **法的リスク回避**: 多くのサイトは利用規約でスクレイピングを禁止している。RSSは公式に提供されている取得手段なので安全
- **Lambda で完結できる**: スクレイピングにはSelenium/Playwrightが必要で、Lambdaのサイズ制限や実行時間制限と相性が悪い。RSSなら軽量なHTTPクライアントだけで済む
- **学習フォーカスの明確化**: 今回の学習目的はインフラとアーキテクチャであり、スクレイピング技術ではない。対応サイトを絞ることで本質に集中できる
- **記事化の安全性**: Zenn記事にした時、規約違反の方法を紹介することにならない

**却下した代替案**:
- 参考記事(ose20さん)と同じくSeleniumでスクレイピング: 学習範囲が広がりすぎる。Phase 4 以降の拡張として検討余地を残す

### ADR-006: なぜローカル開発に LocalStack を使うか

**決定**: ローカル開発環境として LocalStack (Community Edition) を使い、本番と同じ Terraform コードで LocalStack に apply する。

**理由**:
- **料金ゼロ**: 学習中に何度もapply/destroyしても料金が発生しない
- **AWSリソース部分はオフライン開発可能**: ネット不要で反復できる。ただし RSS 取得と Discord 通知の実動作確認にはネット接続が必要
- **本番との差分が小さい**: エンドポイントを差し替えるだけで同じコードが動くので、「ローカルで動いたのに本番で動かない」が減る
- **Terraform の provider 切り替え学習**: 環境ごとに provider 設定を切り替える実務的なパターンを経験できる
- LocalStack Community Edition は Lambda / SQS / DynamoDB / EventBridge を全てサポートしている

**却下した代替案**:
- 本番AWSで直接開発: 料金リスクとdestroy忘れ事故のリスク
- moto などの Python ライブラリ: Terraformから使えない

### ADR-007: なぜ kind + KEDA を後から追加するか

**決定**: Phase 1-5 は純粋にサーバーレスで実装し、Phase 6 で kind + KEDA の k8s consumer を追加する。

**理由**:
- 最初から両方やると挫折する。サーバーレス側で実動作させてから k8s 側を足す方が学習曲線が緩やか
- kind は完全ローカルで料金ゼロ、KEDA の SQS トリガー挙動を手元で観察できる
- 業務で触っている KEDA の挙動を、自分の手で小さく再現することで理解が深まる
- 最終的に「同じGoバイナリがLambdaでもk8s Podでも動く」状態を作れれば、ビルド分離やエントリポイント設計の勘所が学べる

**却下した代替案**:
- EKS: $73/月のコントロールプレーン料金が個人プロジェクトには重い
- Phase 1 から k8s を入れる: 学習範囲が広すぎて挫折リスク高

### ADR-008: なぜ Standard Queue を使うか / FIFO Queue を使わないか

**決定**: SQS Standard Queue を使う。FIFO Queue は使わない。

**理由**:
- 漫画の更新通知に厳密な順序保証は不要。同じ作品の複数話が同時に更新されても、通知順序が前後しても問題ない
- Standard Queue は FIFO より高スループットで、料金も安い
- FIFO Queue は メッセージグループ ID の設計が必要になり、学習コストが上がる
- at-least-once delivery による二重通知は DynamoDB の条件付き書き込みで防ぐ(ADR-011)

**却下した代替案**:
- FIFO Queue: exactly-once delivery で二重通知を防げるが、順序保証が不要なユースケースには過剰。スループット制限(300 msg/s)もある

### ADR-009: SQS + Lambda のリトライ、DLQ、partial batch response をどう設計するか

**決定**: 以下の設計で SQS → Lambda のリトライを管理する。
- SQS visibility timeout: Lambda timeout の6倍以上(例: Lambda 10s → visibility timeout 60s)
- DLQ maxReceiveCount: 5
- Event Source Mapping で `ReportBatchItemFailures` を有効化し、失敗した message ID だけ再処理する

**理由**:
- visibility timeout が Lambda timeout より短いと、Lambda がまだ処理中なのに SQS が同じメッセージを別の invocation に渡してしまう。6倍は AWS の公式推奨
- maxReceiveCount を設定しないと、処理できないメッセージが無限ループする(poison pill 問題)
- ReportBatchItemFailures を使わないと、バッチ内の1件だけ失敗した場合に成功済みメッセージも再処理される

**却下した代替案**:
- バッチサイズ1で常に1件ずつ処理: partial batch response が不要になるが、Lambda invocation 数が増えてコスト増。学習目的としても partial batch response を体験した方がよい

### ADR-010: Discord webhook URL をどう管理するか

**決定**: Phase 4 では Lambda 環境変数で管理し、Phase 5 で SSM Parameter Store (SecureString) に移行する。

**理由**:
- 環境変数は最もシンプルだが、Terraform の state ファイルに平文で残る。個人プロジェクトの初期段階としては許容範囲
- Phase 5 で本番デプロイする際に SSM Parameter Store に移行することで、シークレット管理のベストプラクティスを段階的に学べる
- Secrets Manager は自動ローテーション機能があるが、webhook URL にローテーションは不要。SSM Parameter Store の方がシンプルで無料枠も大きい

**却下した代替案**:
- 最初から Secrets Manager: 学習範囲が広がりすぎる。$0.40/secret/月のコストもある
- ハードコード: 論外。GitHub に push した時点で漏洩する

### ADR-011: DynamoDB で二重通知をどう防ぐか

**決定**: 通知済み管理テーブルを用意し、`attribute_not_exists` の条件付き書き込みで冪等性を保証する。

テーブル設計:
- PK: `manga_id`, SK: `episode_id`
- 属性: `notified_at` (timestamp), `ttl` (DynamoDB TTL で自動削除)

処理フロー:
1. notifier が SQS からメッセージを受け取る
2. DynamoDB に条件付き PutItem (`attribute_not_exists(manga_id)`) ※複合キーテーブルでは PK の存在チェックだけで同一 PK+SK の重複を検知できる
3. 書き込み成功 → Discord に通知
4. ConditionalCheckFailedException → 通知済みなのでスキップ

**理由**:
- SQS Standard Queue は at-least-once delivery なので、同じメッセージが複数回配信される可能性がある
- fetcher 側の差分検知テーブル(PK: `manga_id`, 属性: `latest_episode`)だけでは、notifier 側の二重実行を防げない
- DynamoDB の条件付き書き込みはアトミックなので、複数の Lambda invocation が同時に来ても安全
- TTL で古いレコードを自動削除し、テーブル肥大化を防ぐ

### ADR-012: 監視とアラートをどこまで入れるか

**決定**: Phase 5 で本番デプロイ時に以下の CloudWatch Alarm を作成する。

必須アラーム:
- fetcher Lambda Errors > 0
- notifier Lambda Errors > 0
- DLQ ApproximateNumberOfMessagesVisible > 0
- Lambda Throttles > 0

推奨アラーム:
- SQS ApproximateAgeOfOldestMessage > 300s (メッセージ滞留)

通知先: SNS → メール(個人プロジェクトなので最小限)

**理由**:
- 「デプロイして終わり」ではなく「本番運用している」状態を目指すのが学習目的に合う
- DLQ にメッセージが入っていることに気づかないと、通知が黙って失われる
- Throttles アラームは Lambda の同時実行数制限に引っかかっていないか検知する
- 個人プロジェクトなので PagerDuty 等は不要。メール通知で十分

### ADR-013: なぜ OOP を意識した設計にするか

**決定**: 本プロジェクトのドメイン層は OOP 原則(4本柱、SOLID、継承より合成)を意識して設計・実装する。原則の整理と Go での表現方法は `docs/oop-principles.md` に分離する。

**理由**:
- Ruby キャリア開始時は OOP を十分に理解せず使っていた。Go でも 4 年間 OOP を意識せず書いてきた。本プロジェクトを学習機会として、OOP 原則を体系的に身につける
- Go は継承がなく「合成(has-a)」を言語レベルで強制しており、「継承より合成」という OOP 原則を素直に実践できる
- 原則を意識した設計は、将来ドメインが拡大した時(通知先追加、取得元追加、フィルター条件の追加)の変更容易性に直結する
- 座学メモと設計判断の記録を `docs/oop-principles.md` に切り出すことで、本 spec の肥大化を避け、原則は独立に更新できる

**却下した代替案**:
- Ruby で別プロジェクトを作って OOP 学習: Phase 0 の投資が無駄になる。前職の Go による EC 構築経験の延長として、Go でのドメイン設計の勘所を深める方が実務的
- OOP を意識せず手続き型寄りに書く: 本プロジェクトは学習目的であり、「動くだけ」では学びが薄い
- Ruby 的な OOP 流派(継承ツリー、Mixin、メタプロ)を Go で再現する: 言語の思想に逆らうので採用しない

## 費用の目安

### 本番AWS(常時稼働想定)

| サービス | 用途 | 無料枠 | 想定使用量 | 月額 |
|---------|------|--------|-----------|------|
| Lambda | fetcher (1時間おき) | 100万req/月、40万GB秒 | 720 req/月 | $0 |
| Lambda | notifier (新着時のみ) | 同上 | 数十〜数百 req/月 | $0 |
| SQS | メッセージキュー | 100万req/月 | 数千 req/月 | $0 |
| EventBridge Scheduler | cron | 1400万起動/月まで $1 | 720 起動/月 | $0 |
| DynamoDB | 状態保存 | 25GB、25WCU/25RCU | 数KB、数req/分 | $0 |
| CloudWatch Logs | ログ | 5GB/月 | 数十MB/月 | $0 |
| **合計** | | | | **$0〜1/月** |

**現実的な月額: $0〜1**(ほぼ全て無料枠内に収まる)

追加で発生しうるコスト:
- データ転送: 新着通知を外部(Discord)に送る分。月数MBなので無視できる
- DynamoDB のオンデマンド課金: 無料枠を超えた場合。個人用途なら超えない

### ローカル開発(LocalStack)

**$0**。Docker Desktop と LocalStack Community Edition で完全に無料。

### ローカルk8s(kind + KEDA)

**$0**。kind は Docker 上で動くのでホストマシン以外のコストなし。

### 予算アラートの設定

万が一の事故対策として、AWS Budgets で以下を設定しておく:
- 月額 $5 を超えたらメール通知
- 月額 $10 を超えたらメール通知 + Slack通知(可能なら)

## フェーズ分割

学習を段階的に進めるため、以下のフェーズに分ける。各フェーズは土日半日〜1日で完了するサイズに設計されている。

### Phase 0: 準備

**ゴール**: 開発環境と AWS アカウントの土台を整える。

**やること**:
1. AWS 個人アカウントの作成(既にあればスキップ)
2. IAM Identity Center(SSO)で CLI 認証を設定する(推奨)。学習用途として IAM ユーザーのアクセスキーを使う場合は、MFA・有効期限・最小権限・定期削除を徹底する
3. AWS Budgets で $5/$10 のアラート設定
4. Terraform インストール、バージョン確認
5. LocalStack インストール(Docker Compose で起動できる状態)
6. Go 1.22+ インストール確認
7. GitHub にリポジトリ作成、initial commit

**完了条件**:
- `aws sts get-caller-identity` が成功する
- `terraform version` が 1.6+ を返す
- `docker compose up` で LocalStack が起動する
- `curl http://localhost:4566/_localstack/health` が 200 を返す

### Phase 1: Hello World (LocalStack上)

**ゴール**: LocalStack 上で「Lambda が動いて CloudWatch Logs にログが出る」を達成する。

**やること**:
1. Terraform ディレクトリ構成を作る
   ```
   terraform/
     ├── main.tf
     ├── variables.tf
     ├── outputs.tf
     ├── providers.tf      # LocalStack向けprovider設定
     └── modules/
         └── hello/
             └── main.tf   # Lambda最小構成
   ```
2. Go で最小の Lambda handler を書く(`main.go`)
   - イベントを受け取って "Hello, manga notifier" をログ出力するだけ
3. `GOOS=linux GOARCH=amd64 go build` で Lambda 用バイナリをビルド
4. Terraform で Lambda 関数をデプロイ
5. `aws lambda invoke` (エンドポイント: LocalStack) で手動実行
6. CloudWatch Logs にログが出ることを確認

**完了条件**:
- `terraform apply` が成功する
- Lambda を手動実行して "Hello, manga notifier" がログに出る
- `terraform destroy` で綺麗に消える

**学びポイント**:
- Terraform の provider 設定(LocalStack向けにエンドポイント差し替え)
- Lambda のデプロイパッケージ作成(zip)
- IAM ロールの最小権限設計(CloudWatch Logsへの書き込みだけ)

### Phase 2: EventBridge Scheduler + SQS + 2つのLambda

**ゴール**: fetcher → SQS → notifier の非同期フローを LocalStack で動かす。まだ RSS は取らず、ダミーデータでOK。

**やること**:
1. `modules/fetcher` を作成: EventBridge Schedulerで1分おきに起動するLambda。固定メッセージをSQSにpublishする
2. `modules/queue` を作成: SQSキューと DLQ。以下の設定を明記する:
   - SQS visibility timeout: Lambda timeout の少なくとも6倍(例: notifier timeout 10s → visibility timeout 60s 以上)
   - DLQ maxReceiveCount: 5 以上
3. `modules/notifier` を作成: SQS Event Source Mapping で起動するLambda。受け取ったメッセージをログ出力するだけ
   - Event Source Mapping で `ReportBatchItemFailures` を有効にし、失敗した message ID だけ `batchItemFailures` に入れて返す設計にする(partial batch response)
4. Terraform で全体を apply
5. 1分待って、CloudWatch Logs で以下を確認:
   - fetcher が起動している
   - notifier が SQS からメッセージを受け取っている

**完了条件**:
- EventBridge Scheduler → fetcher → SQS → notifier の流れがログで追える
- SQS のキュー深度を CLI で確認できる
- visibility timeout と DLQ maxReceiveCount が設計通りに設定されている
- `terraform destroy` で全て消える

**学びポイント**:
- EventBridge Scheduler の cron 設定
- SQS の Event Source Mapping
- SQS の visibility timeout と Lambda timeout の関係
- Partial batch response (ReportBatchItemFailures) による部分リトライ
- Lambda → SQS への書き込み権限(IAM)
- SQS → Lambda の読み取り権限(IAM)
- 2つの Lambda の IAM ロールを分ける実践

### Phase 3: DynamoDB と差分検知

**ゴール**: fetcher が DynamoDB に状態を保存し、「前回との差分があった時だけ」SQS に publish するようにする。まだ RSS は使わず、ダミーで「毎回異なる episode 番号を返す関数」でテストする。

**やること**:
1. `modules/storage` を作成: DynamoDB テーブル
   - 差分検知用テーブル: PK: `manga_id`、属性: `latest_episode`, `updated_at`
   - 通知済み管理用テーブル: PK: `manga_id`, SK: `episode_id`、属性: `notified_at`, `ttl`
     - SQS の at-least-once delivery による二重通知を `attribute_not_exists` の条件付き書き込みで防ぐ
2. fetcher Lambda を改修:
   - DynamoDB から前回状態を読み取る
   - ダミー関数から「最新話番号」を取得
   - 前回と異なれば SQS に publish + DynamoDB を更新
   - 同じなら何もしない
3. notifier Lambda を改修:
   - 通知前に DynamoDB の通知済みテーブルを条件付き書き込みでチェック
   - 既に通知済みならスキップ(冪等性保証)
4. 動作確認:
   - 初回実行: DynamoDBに何もないので publish される
   - 2回目(ダミー関数が異なる値を返す設定): publish される
   - 2回目(ダミー関数が同じ値を返す設定): publish されない
   - 同じメッセージを SQS に2回入れても Discord 通知は1回だけ

**完了条件**:
- DynamoDB の中身を AWS CLI (LocalStack向け)で確認できる
- 差分がある時だけ SQS にメッセージが入ることを確認できる
- notifier が正しくメッセージを受け取っている
- 同一エピソードの二重通知が防止されている

**学びポイント**:
- DynamoDB の基本操作(GetItem, PutItem)
- DynamoDB の条件付き書き込み(attribute_not_exists)による冪等性保証
- Lambda の IAM に DynamoDB 権限を追加
- 冪等性の考え方(同じ入力なら同じ結果)
- ADR-004 の「なぜ DynamoDB か」を実感する

### Phase 4: 実際のRSS取得とDiscord通知

**ゴール**: 本物のRSSを取得して、本物のDiscordに通知する。まだLocalStackで動かしていてOK。

**やること**:
1. 追跡したい漫画を決める(少年ジャンプ+のRSSが取れる作品を2〜3作)
   - 例: `https://shonenjumpplus.com/rss/series/{series_id}`
2. fetcher を改修: ダミー関数を `gofeed` での RSS パースに置き換え
3. 追跡対象の作品リストを DynamoDB に登録する方法を決める
   - 案A: Terraform で初期データを投入
   - 案B: CLI スクリプトで手動登録
   - 案C: 環境変数で渡す(一番シンプル、最初はこれで)
4. Discord webhook URL を取得(Discord の適当なサーバーのチャンネル設定から)
5. Discord webhook URL を Lambda の環境変数にセット(Phase 4 では環境変数で管理、Phase 5 で SSM Parameter Store に移行する)
6. notifier を改修: 受け取ったメッセージを整形して Discord に POST
7. LocalStack 環境で実行して Discord に通知が飛ぶことを確認

**完了条件**:
- 実際の漫画の更新を検知できる
- Discord に通知が届く(通知内容: 作品名、最新話タイトル、URL)
- 既読(前回と同じ)の時は通知が飛ばない

**学びポイント**:
- 外部APIとの通信を含む Lambda の実装
- シークレット管理(環境変数 vs Secrets Manager の選択)
- RSS フィードのパースとドメインモデルへのマッピング
- Discord webhook の基本

**ドメイン設計のヒント**:
ose20さん記事のアイデアを拝借して、以下のドメインモデルを考える:
- `Manga`: title, portal, crawl_url, public_url
- `MangaEpisode`: episode_title, published_at
- `MangaPortal`: enum的に JumpPlus, ComicDays など
- Clean Architecture 的に `internal/domain`, `internal/usecase`, `internal/infra` に分ける

### Phase 5: 本番AWSへのデプロイ

**ゴール**: LocalStack で動いたものを本番 AWS に展開する。

**やること**:
1. Terraform の provider 設定を環境切り替えできるようにする
   - `terraform/envs/local/` と `terraform/envs/prod/` に分ける
   - または `terraform workspace` を使う
2. Terraform state を S3 バックエンドに移す(本番用)
   - S3バケット + DynamoDBロックテーブルを手動で作成(chicken-and-egg問題対応)
3. 本番 AWS に apply
4. 1時間後に Discord に通知が来ることを確認
5. AWS Console で各リソースの動作を目視確認

**完了条件**:
- 本番 AWS で fetcher が定期実行されている
- 新着があれば Discord に通知が来る
- AWS Budgets のアラートが設定されている
- CloudWatch Logs で各 Lambda のログが見られる
- CloudWatch Alarm が作成されている(fetcher/notifier Errors, DLQ メッセージ数, Lambda Throttles, SQS メッセージ滞留時間)
- Discord webhook URL が SSM Parameter Store に移行されている

**学びポイント**:
- Terraform の環境分離パターン
- Terraform state のリモート管理
- 本番とローカルの差分をどう吸収するか
- 本番で初めて発見される問題(LocalStackにない機能、IAM権限の違い等)

### Phase 6: ローカルk8s (kind + KEDA) 版 consumer の追加

**ゴール**: notifier と同じ処理を kind 上の Go Pod として動かし、KEDA で SQS トリガーの scale を観察する。

**やること**:
1. kind でローカルクラスタを作成
2. 同じ Go コードから k8s 版エントリポイントを作る(Lambda handler とは別の `cmd/worker/main.go`)
3. Dockerfile を書いて Docker イメージをビルド、kind にロード
4. k8s manifest を書く:
   - Deployment(replicas: 0 から始める)
   - ServiceAccount
   - Secret (SQSエンドポイント、Discord webhook URL)
5. KEDA をインストール
6. ScaledObject を書く: SQS キューの深度が1以上になったら Pod を起こす
7. LocalStackのSQSにメッセージを入れてPodが起動する様子を観察
   - `kubectl get pods -w` で watch しながら

**完了条件**:
- kind クラスタが起動している
- KEDA がインストールされている
- SQSにメッセージが入ると Pod が scale up する
- メッセージを処理し終わると Pod が scale down する(scale-to-zero)
- Discord に通知が届く(Lambda版と同じ挙動)

**学びポイント**:
- kind の使い方(クラスタ作成、イメージロード)
- 同じドメインロジックをLambda handlerとk8s workerの両方から呼べる設計
- KEDA の ScaledObject と SQS トリガー
- k8s manifest の基本(Deployment, ServiceAccount, Secret)
- scale-to-zero の実際の挙動観察

**業務との接続**:
このフェーズで体験する「KEDA + SQS で scale する」という挙動は業務の本番運用と同じパターン。自分の手で作ることで、業務コードがなぜあの形なのかが腑に落ちる。

### Phase 7: 記事化

**ゴール**: Zenn記事を執筆して、学びを言語化する。

**記事案**:
1. 「Go + サーバーレスで漫画更新通知botを作った - 業務のhistory_jobパターンを個人プロジェクトに適用した話」
2. 「LocalStackとTerraformで料金ゼロのAWS学習環境を作る」
3. 「同じGoコードをLambdaとk8s Podの両方で動かす - kind + KEDA で scale-to-zero を体験する」

## ディレクトリ構成(最終形)

```
manga-notifier/
├── README.md
├── go.mod
├── go.sum
├── cmd/
│   ├── fetcher/       # Lambda handler: RSS取得 + SQS publish
│   │   └── main.go
│   ├── notifier/      # Lambda handler: Discord通知
│   │   └── main.go
│   └── worker/        # k8s worker エントリポイント
│       └── main.go
├── internal/
│   ├── domain/        # ドメインモデル(Manga, MangaEpisode, MangaPortal)
│   ├── usecase/       # fetch usecase, notify usecase
│   └── infra/         # DynamoDB実装, SQS実装, Discord実装, RSS実装
├── terraform/
│   ├── envs/
│   │   ├── local/     # LocalStack向け
│   │   └── prod/      # 本番AWS向け
│   └── modules/
│       ├── fetcher/
│       ├── notifier/
│       ├── queue/
│       └── storage/
├── k8s/
│   ├── base/          # Deployment, ServiceAccount, Secret
│   └── keda/          # ScaledObject
├── docker/
│   ├── Dockerfile.worker
│   └── compose.yml    # LocalStack起動用
└── docs/
    ├── ADR.md         # このファイルのADRセクションをここに分離してもよい
    └── phases/        # 各フェーズの振り返りメモ
```

## 注意事項・落とし穴

- **LocalStackのLambdaは本番と挙動が微妙に違う**: cold start、IAM評価のタイミングなど。Phase 5 で本番にデプロイした時に差分が出ることを覚悟しておく
- **LocalStack Community Edition の制限**: 一部のサービス(EventBridge Schedulerの細かい機能など)は Pro 版でしか動かないことがある。エラーが出たら代替手段(CloudWatch Events)を検討
- **DynamoDBの書き込み単位**: 個人用途では絶対超えないが、ループ処理などで意図せず大量書き込みを発行しないように注意
- **Discord webhook のレート制限**: 1秒間に5リクエストが上限。notifier の同時実行数を1に制限することを推奨
- **destroy忘れ事故の防止**: 本番AWSは使い終わったら必ず `terraform destroy` する癖をつける。Budget アラートは最後の砦
- **著作権への配慮**: 表紙画像をLambdaからダウンロードして再配信するのはNG。Discord通知ではURLを埋め込む形にとどめる

## 成功の定義

このプロジェクトは以下が達成できたら成功:

1. ✅ 本番AWS環境で漫画の更新が自動検知され、Discordに通知が来る
2. ✅ 同じTerraformコードがLocalStackでも動き、ローカル開発が完全無料
3. ✅ kind + KEDA で k8s 版 consumer が動作し、Lambda版と同じ挙動を示す
4. ✅ Zenn記事を1本以上公開した
5. ✅ 月額費用が$5を超えていない
6. ✅ 業務で触っている KEDA / IRSA / history_job パターンへの理解が深まった実感がある

