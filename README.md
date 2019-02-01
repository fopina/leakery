docker run --rm -it -p 28000:27017 -v $(pwd)/mongodb:/data/db --name leakery mongo:3.6


docker run --rm -it -p 28000:27017 -v $(pwd)/mongodb:/data/db --name leakery mangoraft/mongodb-arm
