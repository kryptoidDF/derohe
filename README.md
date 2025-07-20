# DEROHE (Beta Fork)

🚧 **Pre-release Beta**  
This is a **beta fork** of the original [civilware/derohe](https://github.com/civilware/derohe) project. Although it is completely stable, it is provided as-is for testing and development purposes only.

---

## 🔧 Changes in This Fork

- **Fix for randomness reuse** (Stops brute forcing of transaction payloads)
- **Default ring size** has been set to **32**
- The `--remote` flag now points to the **DERO Foundation public node**

---

## 📦 Original Project

This repository is forked from [civilware/derohe](https://github.com/civilware/derohe) **DEV BRANCH**, an implementation of the DERO Homomorphic Encryption blockchain daemon.

Please refer to the upstream project for full documentation and historical context.

---

## ⚠️ Disclaimer

This is **experimental software**. Use at your own risk. Do not use with mainnet funds unless you understand the implications of running beta software.

---

## 🧪 Testing

To launch the wallet type the following into your command line :

**Linux :** ./dero-cli-wallet-linux-amd64 --remote
**Windows :** - dero-cli-wallet-windows-amd64 --remote
**MacOS :** ./dero-cli-wallet-darwin-universal --remote