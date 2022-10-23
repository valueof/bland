# Bland

It's like del.icio.us of the past but you can host it yourself!

![Screenshot of Bland, from October 22nd, 2022](meta/screenshot%2010-22-22.png)

## Quick Start

Create a new build:
```sh
make
```

Switch to the new directory and setup the database. This will create a new SQLite database file (bland.db) including all the necessary tables:
```sh
cd ./build
./bland -db bland.db -setup
```

Now you can run the server:
```sh
./bland -db bland.db -addr localhost:9999
```

## Development
To start a development server run this:
```sh
go run . -dev -db bland.db -addr localhost:9999
```

If you have [nodemon](https://nodemon.io/) installed you can watch for changes and reload the server automatically:
```sh
nodemon --exec go run . -dev -db bland.db -addr localhost:9999 --signal SIGTERM --ext html,go
```

## Optional
### Import from Pinboard
If you, like me, have a JSON file with data from Pinboard you can import it into your database while setting it up:
```sh
./bland -db bland.db -setup -seed /path/to/pinboard_export.json
```

### Run Bland in the background
For longer running instances I highly recommend running Bland as a background service and putting it behind a reverse proxy server such as [Nginx](https://www.nginx.com/) or [Caddy](https://caddyserver.com).

#### On Ubuntu Linux
First, create a new file in the `/lib/systemd/system` directory named `bland.service` and make it something like this (this assumes your `build` directory from the Quick Start section above is in `/home/anton/srv/bland`):
```
[Unit]
Description=bland

[Service]
Type=simple
Restart=always
RestartSec=5s
WorkingDirectory=/home/anton/srv/bland
ExecStart=/home/anton/srv/bland/bland -addr localhost:9999 -db /home/anton/srv/bland/bland.db
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Save this file. You can now start Bland in the background by running:
```sh
sudo service bland start
```

You can now put Bland behind a reverse proxy server. If you have [Caddy installed](https://caddyserver.com/docs/getting-started) it's as simple as adding this to your Caddyfile:
```
myblanddomain.ts.net {
    reverse_proxy localhost:9999
}
```

If you want to use Nginx, check out this excellent guide from DigitalOcean: [How To Deploy a Go Web Application Using Nginx on Ubuntu](https://www.digitalocean.com/community/tutorials/how-to-deploy-a-go-web-application-using-nginx-on-ubuntu-18-04).

Enjoy!