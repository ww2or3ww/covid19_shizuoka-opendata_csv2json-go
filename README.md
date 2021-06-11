# covid19_shizuoka-opendata_csv2json-go

## 開発環境

```
$ go version
go version go1.13.4 linux/amd64
```

## ディレクトリ構成

```
|--bin
|  |--main
|--src
|  |--main
|  |  |--go.mod
|  |  |--go.sum
|  |  |--main.go
|  |  |--utils
|  |  |  |--logging.go
```

## 初期セットアップ

```
$ cd src/app
$ go mod init app
```

## デバッグ実行

```
$ go run main.go
```


## ビルド & パッケージング & デプロイ

```
$ GOOS=linux GOARCH=amd64 go build -o ../../bin/main main.go
$ (cd ../../bin && zip -r ../lambda-package.zip *)
$ aws lambda update-function-code --function-name covid19_hamamatsu_csv2json_go --zip-file fileb://../../lambda-package.zip
```

## テスト

```
$ go test -cover
```

