build:
	go build -o bin/
build-prod:
	go build -ldflags "-s -w" -o bin/
run:
	go run .
clean:
	rm -rf bin/

build-mac: clean
	mkdir bin
	fyne package -os darwin --release
	mv Pinecone.app bin/
  	cp -r data/ bin/data/

build-linux: clean
	mkdir bin
	fyne package -os linux --release
	mv Pinecone.tar.xz bin/
 	cp -r data/ bin/data/

build-win: clean
	mkdir bin
	fyne package -os windows --release
	mv Pinecone.exe bin/
	cp -r data/ bin/data/
 