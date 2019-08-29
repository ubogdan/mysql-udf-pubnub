
udf_file="pubnub_udf.so"

#all: build install

build:
	GOPATH=`pwd` go build -buildmode=c-shared -ldflags "-s -w" -o ${udf_file}

install:
	/etc/init.d/mysql stop
	install -o root -g root ${udf_file} -t `mysql_config  --plugindir`
	/etc/init.d/mysql start