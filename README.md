# happygeocoding

「[陛下が発見したタヌキ](http://www.kunaicho.go.jp/page/ronbun/show/2)を見られるモバイルアプリ」みたいなやーつ（誰得）を想定した、バックエンドAPI

## how to use

```shell
# ローカルサーバー起動
dev_appserver.py --port=8081 --admin_port=8001  server/app.yaml 

# テストデータ投入(初回のみ)
go run tool/post.go 20
```

## link

[qiita](https://qiita.com/tarokamikaze/items/4f5327bccc7b166397ec)
