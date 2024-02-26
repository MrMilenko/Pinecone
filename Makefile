build:
	go build -o bin/
build-prod:
	go build -ldflags "-s -w" -o bin/
run:
	go run .
tests:
	go test ./...
clean:
	rm -rf bin/ muscurdi_db/

build-mac: clean
	mkdir bin
	fyne package -os darwin -icon assets/logo.png --release
	mv Pinecone.app bin/
  	cp -r data/ bin/data/

build-linux: clean
	mkdir bin
	fyne package -os linux -icon assets/logo.png --release
	mv Pinecone.tar.xz bin/
 	cp -r data/ bin/data/

build-win: clean
	mkdir bin
	fyne package -os windows -icon assets/logo.png --release
	mv Pinecone.exe bin/
	cp -r data/ bin/data/
 