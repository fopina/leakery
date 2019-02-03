### mysql

docker run --rm --name leakery -v $(pwd)/mysql:/var/lib/mysql -e MYSQL_DATABASE=test -e MYSQL_ROOT_PASSWORD=password -p3306:3306 mysql:5.5


### mongo

docker run --rm -it -p 28000:27017 -v $(pwd)/mongodb:/data/db --name leakery mongo:3.6
docker run --rm -it -p 28000:27017 -v $(pwd)/mongodb:/data/db --name leakery mangoraft/mongodb-arm
