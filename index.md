---

layout: col-sidebar
title: OWASP KubeFIM
tags: kubefim kubernetes FIM
level: 2
type: Code
pitch: OWASP KubeFIM is an eBPF based Kubernetes File Integrity Monitoring solution.

---

OWASP KubeFIM is an eBPF based Kubernetes FIM solutions which detects the file system changes in the Kubernetes cluster. The specifications are below,
* Detect the change of file system in the Kubernetes Nodes.
* Detect the suspecious malware activites inside the Pod.

[![alt text](/assets/images/kubefim.png)](https://owasp.org/www-project-kubefim/)


### Background
With the adoption of the Kubernetes to run the workloads, new threats are coming everyday. It includes the malicious containers running inside the cluster silently and excuting malicious code. KubeFIM detects any suspicious changes in file system and alert the system admins & security team.

OWASP KubeFIM was created for this purpose.

OWASP KubeFIM is backed by the [OWASP Foundation](https://owasp.org)