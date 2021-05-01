# kubectl-cf

Faster way to switch between kubeconfigs (not contexts).

![demo.gif](https://github.com/spongeprojects/kubectl-cf/blob/main/assets/demo.gif?raw=true)

```
Usage of kubectl-cf:

  cf           Select kubeconfig interactively
  cf [config]  Select kubeconfig directly
  cf -         Switch to the previous kubeconfig
```

This tool is designed to switch between kubeconfigs, if you want to switch between context within a single kubeconfig (
or multiple kubeconfigs), you should use https://github.com/ahmetb/kubectx instead.

### Installation

#### Install Manually

First, download tar file from the release page: https://github.com/spongeprojects/kubectl-cf/releases.

After unzip the tar file, you'll get the executable file named `kubectl-cf`.

Put it in any place you want as long as it's in your `PATH`. It can be called directly by typing `kubectl-cf`, or as
a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) `kubectl cf`, because it has the
prefix `kubectl-`.

You can also rename it to any name you want, or create a symlink to it, with a shorter name, like `cf`.

### TODO (PR are welcomed)

- Auto completion;
- [krew](https://krew.sigs.k8s.io/) integration;
- Tests;
