# Gocryptfs keystore

One of the methods to store private keys of watchtowers and operators is 
gocryptfs. It store keys in encrypted format when the key is not in use 
(data at rest).

## Gocryptfs: store keys in encrypted format
Gocryptfs provide an encrypted file system. It uses FUSE (Filesystem in 
userpace) to mount encrypted directory which are decrypted when the 
files system is mounted.

### Pre-requisite
Install gocryptfs

```bash
sudo apt install gocryptfs
```

