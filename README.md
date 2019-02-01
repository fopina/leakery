docker run --rm --name leakery -v $(pwd)/mysql:/var/lib/mysql -e MYSQL_DATABASE=test -e MYSQL_ROOT_PASSWORD=password -p3306:3306 mysql:5.5
