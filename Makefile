arm:
	env GOOS=linux GOARCH=arm go build -o leakery-arm withbolt2.go

mongo:
	go run withmongo.go -import $(HOME)/test

mysql:
	go run withmysql.go -import $(HOME)/test

sqlite:
	rm -f sqlite.db
	go run withsqlite.go -import $(HOME)/test

sqlite2:
	rm -f sqlite.db
	go run withsqlite2.go -import $(HOME)/test

sqlitetest:
	sqlite3 sqlite.db 'select count(*) from leaks';

redis:
	go run withredis.go -import $(HOME)/test -debug

bolt:
	rm -f bolt.db
	go run withbolt.go -import $(HOME)/test

bolt2:
	rm -f bolt.db
	go run withbolt2.go -import $(HOME)/test

bolt3:
	rm -f bolt.db
	go run withbolt3.go -import $(HOME)/test

clearmysql:
	docker rm -f mysql-leaker || echo 1
	docker run --rm --name mysql-leaker -e MYSQL_DATABASE=test -e MYSQL_ROOT_PASSWORD=password -p3306:3306 mysql:5.5

clearmongo:
	docker rm -f mongo-leaker || echo 1
	docker run --rm -it --name mongo-leaker -p 28000:27017 mongo:3.6

clearredis:
	docker rm -f redis-leaker || echo 1
	docker run --rm -it --name redis-leaker -p 6379:6379 redis:4-alpine