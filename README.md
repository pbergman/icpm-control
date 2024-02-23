## ICPM Control

A simple server/client wich was inspired by knockd but uses signed ipcm packages to execute predefined commands on the server and gives a bit more flexibility for the things to execute. 

### Server

The server uses pcap to listen for incoming messages, so it can be used behind a firewall. 

Executing of scripts is done by opening a shell (default bash but can be changed) and writing the exec to the stdin. 

It will be processed by the  [text/template](https://pkg.go.dev/text/template) so that should give you more flexibility for creating dynamic scripts and there is also possible to define global block which can be reused.

So you should be able to do something like this:

```

block.echo = "echo \"hello {{ . }}\""

[[script]]
code = 2
exec = [\
    'set -x'
    '{{ template "echo" "foo" }}',
    '{{ template "echo" "bar" }}',
]

```

Which should print something like this in the logs when called. 

```
..] + echo 'hello "foo"'
..] + echo 'hello "bar"'
```

More info about hte options can be seen [read here](/server/example.conf).


This project was made to run on an [Edge Router](https://eu.store.ui.com/eu/en/collections/uisp-wired-advanced-routing-compact-poe/products/er-x) so to do a cross compile you could use the following commands.

```
make docker-start-mipsel
make build-server-mipsel
make docker-stop-mipsel
```


### Client

build executable: 

```
make build-client
```

Because the client uses raw socket to open connection it will need sudo, to give user permission you can do:

```
sudo setcap cap_net_raw+ep /usr/local/bin/icpm-control
```

On first run it will generate key pairs (will be saved in `~/.config/icpm-control`) en output the public key in the og file which should be used in the server config.

### Server Example

the following example should add your ip address to the ADDR_GROUP_WHITELIST which can be used in a rule. 

```
interface = "switch0"

clients = [
    "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
]

block.vyatta_exec = """
/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper begin
/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper {{ . }}
/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper commit
/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper save
/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper end
"""

[[script]]
var.address_group = "ADDR_GROUP_WHITELIST"
user="ubnt"
group="vyattacfg"
code=0x01
exec = """
{{ template "vyatta_exec" printf "set firewall group address-group %s address \"%s\"" .address_group .src_ip }}
cat << EOF | /usr/bin/at now + 30 minute
  {{ template "vyatta_exec" printf "delete firewall group address-group %s address \"%s\"" .address_group .src_ip }}
EOF
"""
```

for a more verbose output in the logs you could use the following

```
# output excuted command for debuggin
set -x
{{ template "vyatta_exec" printf "set firewall group address-group %s address \"%s\"" .address_group .src_ip }}
# debug output to see if we are allowed
sudo ipset test {{ .address_group }} {{ .src_ip }}
cat << EOF | /usr/bin/at now + 30 minute
  {{ template "vyatta_exec" printf "delete firewall group address-group %s address \"%s\"" .address_group .src_ip }}
EOF
```