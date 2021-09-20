# covid19_shizuoka-opendata_csv2json-go

「静岡県ふじのくにオープンデータカタログ」で公開されているCSVデータを  
「新型コロナウイルス感染症対策サイト」で利用しているdata.jsonへ変換するプロジェクトです。  
(* Amazon API Gatewayから呼び出される AWS Lambda にデプロイして実行される事を想定した実装となっています。)

## サポートデータ

| 名称 | COVID-19サイトのデータタイプKEY | オープンデータカタログID(浜松市用) |
| --- | --- | --- |
| 検査陽性者の状況 | main_summary | 5ab47071-3651-457c-ae2b-bfb8fdbe1af1 & 92f9ebcd-a3f1-4d5d-899b-d69214294a45 |
| 陽性患者の属性 | patients | 5ab47071-3651-457c-ae2b-bfb8fdbe1af1 |
| 陽性患者数 | patients_summary | 5ab47071-3651-457c-ae2b-bfb8fdbe1af1 |
| 検査実施人数 | inspection_persons| d4827176-d887-412a-9344-f84f161786a2 |
| 新型コロナウイルス感染症に関する相談件数| contacts | 1b57f2c0-081e-4664-ba28-9cce56d0b314 |

(* [検査陽性者の状況] は、死亡者のカウントのために、2つのCSVを参照しています。)

## クエリパラメータ引数について

| key | value |
| --- | --- |
| type | GraphType-key:API-IDの配列 |

```bash
example
?type=main_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,main_summary:92f9ebcd-a3f1-4d5d-899b-d69214294a45,patients:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,patients_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1,inspection_persons:d4827176-d887-412a-9344-f84f161786a2,contacts:1b57f2c0-081e-4664-ba28-9cce56d0b314
```

## Deploy to Lambda (zip)

```bash
commands
$ pwd
{workspaceRoot}/src/app

$ GOOS=linux GOARCH=amd64 go build -o ../../bin/main main.go
$ (cd ../../bin && zip -r ../lambda-package.zip *)
$ aws lambda update-function-code --function-name ${LAMBDA_FUNCTION_NAME} --zip-file fileb://../../lambda-package.zip
```

## Deploy to Lambda (Container image)

```bash
commands
$ pwd
{workspaceRoot}
$ aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-1.amazonaws.com

$ docker build -f Dockerfile.release -t csv2json-release .
$ docker tag csv2json-release:latest ${AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-1.amazonaws.com/csv2json:latest
$ docker push ${AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-1.amazonaws.com/csv2json:latest
$ aws lambda update-function-code --function-name ${LAMBDA_FUNCTION_NAME} --image-uri ${AWS_ACCOUNT_ID}.dkr.ecr.ap-northeast-1.amazonaws.com/csv2json:latest
```

## Docker for local

```bash
commands
$ pwd
{workspaceRoot}

# build image & run container
$ docker build -f Dockerfile.debug -t csv2json-debug .
$ docker run --rm -p 9000:8080 csv2json-debug:latest /main

# request for test
$ curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{}' -o ret.json

# request for test (with query parameters)
$ curl -XPOST  \
    "http://localhost:9000/2015-03-31/functions/function/invocations"  \
    -d '{ "queryStringParameters" : { "type" : "main_summary:5ab47071-3651-457c-ae2b-bfb8fdbe1af1" } }' \
    -o ret.json
```

## ユニットテスト

```bash
commands
$ pwd
{workspaceRoot}/src/app

$ go test
```
