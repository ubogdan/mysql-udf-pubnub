# Pubnub.com Mysql UDF

We are not using this plugin any more. The plugin is good/working state. 

1. edit udf_pubnub.go and setup corret publish(pubKey), subscribe(subKey), security(secKey).
```
const (
	pubKey = "pub-c-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	subKey = "sub-c-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	secKey = "sec-c-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
)
```
2. Build plugin
```
make build 
```

3. Upload the plugin into server by ftp and copy the file in the mysql plugin path 
```mysql
mysql> show variables like 'plugin_dir';
+---------------+--------------------------+
| Variable_name | Value                    |
+---------------+--------------------------+
| plugin_dir    | /usr/lib64/mysql/plugin/ |
+---------------+--------------------------+
1 row in set (0.00 sec)
```

4. Install the plugin 
```mysql
DROP FUNCTION IF EXISTS pubnub_grant;
CREATE FUNCTION pubnub_publish RETURNS INT SONAME 'pubnub_udf.so'
DROP FUNCTION IF EXISTS pubnub_publish;
CREATE FUNCTION pubnub_grant RETURNS INT SONAME 'pubnub_udf.so';
```
