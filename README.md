# kubectl-upgrade-readiness

A `kubectl` plugin to validate Kubernetes cluster upgrade readiness.  
It checks for critical preconditions like kubelet version uniformity and control plane addon compatibility — before you initiate a control plane upgrade.

---

![Go](https://img.shields.io/badge/Go-1.21%2B-blue?logo=go)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.25%2B-326ce5?logo=kubernetes)
![Plugin](https://img.shields.io/badge/kubectl-plugin-lightgrey?logo=kubernetes)

---

## ✨ Features

- Detects kubelet version skew across nodes
- Validates `kube-proxy`, `CoreDNS`, and `metrics-server` compatibility
- CLI-native output with actionable `[OK]` / `[!]` flags
- Designed for CI/CD and manual upgrade workflows

---

## 📦 Installation

```bash
go install github.com/marulkar/kubectl-upgrade-readiness@latest
````

After build, place the binary in your `$PATH`:

```bash
mv kubectl-upgrade_readiness /usr/local/bin/
```

> ✅ Your plugin can now be used via `kubectl upgrade_readiness`.

---

## 🚀 Usage

```bash
kubectl upgrade_readiness --target-version=v1.28
```

Sample Output:

```
kubectl-upgrade-readiness: MVP Check (target: v1.28)

Kubelet Version Uniformity Check:
  node-a => kubelet v1.28.5
  node-b => kubelet v1.28.5
  [OK] All kubelets match the target version.

Control Plane Addon Compatibility:
  coredns-xxx => coredns:v1.9.3
    [!] Consider upgrading CoreDNS to v1.10+ for v1.28 compatibility
  metrics-server-yyy => metrics-server:v0.5.0
    [!] Metrics-server is outdated for v1.28
```

---

## 🛡️ Permissions

Requires only read access:

* `get/list` on nodes and pods

---

## 🧭 Roadmap

* 🔜 API deprecation checks (live detection of deprecated resources)
* JSON output for CI pipelines
* Krew packaging


---

## 📄 License

MIT
