arm:
	env GOOS=linux GOARCH=arm go build -o leakery-arm

mongo:
	go run withmongo.go -import $(HOME)/test

mysql:
	go run withmysql.go -import $(HOME)/test

sqlite:
	go run withsqlite.go -import $(HOME)/test -debug

bolt:
	rm -f bolt.db
	go run withbolt.go -import $(HOME)/test

clearmysql:
	docker rm -f mysql-leaker || echo 1
	docker run --rm --name mysql-leaker -e MYSQL_DATABASE=test -e MYSQL_ROOT_PASSWORD=password -p3306:3306 mysql:5.5

clearmongo:
	docker rm -f mongo-leaker || echo 1
	docker run --rm -it --name mongo-leaker -p 28000:27017 mongo:3.6
