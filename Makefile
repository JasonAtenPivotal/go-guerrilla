all:
	cd gomailclient; go build; go install
	go build
	go install
	# if you need certs see: make cert

cert:
	# will overwrite key.pem and cert.pem in your current directory.
	cd generate_cert; go build; go install;
	./generate_cert/generate_cert

debug:
	go build  -gcflags "-N -l"
	go install

clean:
	rm -f  *~ *.o generate_cert/generate_cert  go-guerrilla  go-guerrilla.test  gomailclient/gomailclient

testbuild:
	go test -c -gcflags "-N -l" -v

test:
	# must have run make cert first
	go test -v

