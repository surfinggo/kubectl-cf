# kubectl-cf

Faster way to switch between kubeconfigs (not contexts).

![demo.gif](https://github.com/spongeprojects/kubectl-cf/blob/main/assets/demo.gif?raw=true)

```
Usage of kubectl-cf:

  cf           Select kubeconfig interactively
  cf [config]  Select kubeconfig directly
  cf -         Switch to the previous kubeconfig
```

This tool is designed to switch between kubeconfigs, if you want to switch between context within a single kubeconfig (or multiple kubeconfigs), you should use https://github.com/ahmetb/kubectx instead.
