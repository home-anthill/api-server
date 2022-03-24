# Certbot on Ubuntu 20.04

For more info: https://certbot.eff.org/instructions?ws=nginx&os=ubuntufocal

1. Install and update snap

```bash
sudo apt update
sudo apt install snapd

sudo snap install core; sudo snap refresh core
```

If you are already using certbot via apt, remove it:

```bash
sudo apt-get remove certbot
```

2. Install Certbot

```bash
sudo snap install --classic certbot
```

3. Prepare the Certbot command

```bash
sudo ln -s /snap/bin/certbot /usr/bin/certbot
```

4. Run Certbot

```bash
sudo certbot --nginx
```

Instead, If you don't want to update nginx config run:

```bash
sudo certbot certonly --nginx
```

5. Test automatic renewal

```bash
sudo certbot renew --dry-run
```