
Create a service file 

```
sudo vi /etc/systemd/system/icpm-control.servic
```

Paste following to the file

```
[Unit]
Description=icpm controll package listerner 
 
[Service]
User=root            
ExecStart=/usr/bin/icpm-control                  
Restart=always                               
RestartSec=3          
              
[Install]   
WantedBy=multi-user.target

```

Execute following commands to reload daemon to load new config, enable to run on startup and start the service

```
sudo systemctl daemon-reload
sudo systemctl enable icpm-control
sudo systemctl start icpm-control
sudo systemctl status icpm-control
```

The server will print all logs to sdout but the edgerouter > 2.09 has journalctl disabled by default. So i changed the config to following: 

```
~:# show system systemd journal
 max-retention 60
 runtime-max-use 32
 storage volatile
```

and now i could debug logs with

```
journalctl -u icpm-control -f
```

