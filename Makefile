wa:
	go build
install:
	cp -v wa /usr/local/bin/
	chmod a+xr /usr/local/bin/wa
uninstall:
	rm -v /usr/local/bin/wa
