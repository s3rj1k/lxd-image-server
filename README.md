# lxd-image-server
Dumb LXD Image server for `lxc image import URL`

It is intended to serve template images prepared as `Unified tarball` as defined in: 

https://lxd.readthedocs.io/en/stable-2.0/image-handling/

You will need to configure reverse proxy load banancer as as nginx in from of this application to use it for SSL termination.

Example nginx configuration for proxy pass:
```
  location / {
    proxy_buffering off;
    proxy_http_version 1.1;
    proxy_set_header X-Forwarded-Host $server_name;
    proxy_pass http://127.0.0.1:8000;
  }
```
