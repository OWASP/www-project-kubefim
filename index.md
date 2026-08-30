---
layout: col-sidebar
title: OWASP KubeFIM
tags: kubefim kubernetes eBPF FIM runtime-security
level: 2
type: Code
pitch: Kernel-level file and process visibility for Kubernetes.
---

<style>
  body.col-sidebar .page-title { display: none; }

  .kf-page {
    --kf-ink: #10233f;
    --kf-muted: #52657d;
    --kf-blue: #0867e8;
    --kf-blue-dark: #0754bd;
    --kf-cyan: #19b9d1;
    --kf-line: #dce7f3;
    --kf-soft: #f3f8fd;
    --kf-white: #ffffff;
    color: var(--kf-ink);
    font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    font-size: 16px;
    line-height: 1.65;
  }

  .kf-page *, .kf-page *::before, .kf-page *::after { box-sizing: border-box; }
  .kf-page h1, .kf-page h2, .kf-page h3, .kf-page p { margin-top: 0; }
  .kf-page h1, .kf-page h2, .kf-page h3 { color: var(--kf-ink); letter-spacing: -.035em; }
  .kf-page a { color: var(--kf-blue-dark); }

  .kf-shell {
    width: min(1120px, 100%);
    margin: 0 auto;
  }

  .kf-hero {
    position: relative;
    overflow: hidden;
    margin-bottom: 22px;
    padding: clamp(44px, 7vw, 88px) clamp(24px, 6vw, 72px);
    border: 1px solid #d9e9fb;
    border-radius: 28px;
    background:
      radial-gradient(circle at 88% 12%, rgba(25, 185, 209, .19), transparent 31%),
      radial-gradient(circle at 72% 88%, rgba(8, 103, 232, .12), transparent 37%),
      linear-gradient(145deg, #fff 10%, #f5faff 66%, #edf7ff);
    box-shadow: 0 24px 70px rgba(20, 69, 121, .10);
  }

  .kf-hero-grid {
    position: relative;
    z-index: 1;
    display: grid;
    grid-template-columns: minmax(0, 1.3fr) minmax(260px, .7fr);
    gap: clamp(28px, 6vw, 72px);
    align-items: center;
  }

  .kf-kicker, .kf-section-kicker {
    display: inline-flex;
    align-items: center;
    gap: 9px;
    margin-bottom: 18px;
    color: #165a9f;
    font-size: 12px;
    font-weight: 750;
    letter-spacing: .13em;
    text-transform: uppercase;
  }

  .kf-kicker::before {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #20b789;
    box-shadow: 0 0 0 6px rgba(32, 183, 137, .12);
    content: "";
  }

  .kf-hero h1 {
    max-width: 760px;
    margin-bottom: 22px;
    font-size: clamp(42px, 6vw, 72px);
    font-weight: 760;
    line-height: .99;
  }

  .kf-hero h1 span {
    color: var(--kf-blue);
  }

  .kf-lede {
    max-width: 650px;
    margin-bottom: 30px !important;
    color: var(--kf-muted);
    font-size: clamp(17px, 2vw, 20px);
    line-height: 1.6;
  }

  .kf-actions { display: flex; flex-wrap: wrap; gap: 12px; }
  .kf-button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-height: 48px;
    padding: 0 20px;
    border: 1px solid #b8cce1;
    border-radius: 12px;
    background: rgba(255,255,255,.78);
    color: var(--kf-ink) !important;
    font-size: 14px;
    font-weight: 700;
    text-decoration: none !important;
    transition: transform .18s ease, box-shadow .18s ease, border-color .18s ease;
  }
  .kf-button:hover { transform: translateY(-2px); border-color: #83a9cf; box-shadow: 0 10px 24px rgba(16, 35, 63, .10); }
  .kf-button-primary { border-color: var(--kf-blue); background: var(--kf-blue); color: #fff !important; }
  .kf-button-primary:hover { border-color: var(--kf-blue-dark); background: var(--kf-blue-dark); }
  .kf-button:focus-visible { outline: 3px solid rgba(8, 103, 232, .28); outline-offset: 3px; }

  .kf-signal-stage {
    position: relative;
    min-height: 330px;
    display: grid;
    place-items: center;
  }
  .kf-orbit {
    position: absolute;
    width: 295px;
    height: 295px;
    border: 1px solid rgba(8, 103, 232, .15);
    border-radius: 50%;
  }
  .kf-orbit::before, .kf-orbit::after {
    position: absolute;
    border: 1px solid rgba(8, 103, 232, .12);
    border-radius: 50%;
    content: "";
  }
  .kf-orbit::before { inset: 30px; }
  .kf-orbit::after { inset: 66px; }
  .kf-logo {
    position: relative;
    z-index: 1;
    width: 220px;
    max-width: 74%;
    filter: drop-shadow(0 20px 24px rgba(0, 77, 166, .14));
  }
  .kf-event-card {
    position: absolute;
    z-index: 2;
    right: -8px;
    bottom: 6px;
    width: min(260px, 88%);
    padding: 16px;
    border: 1px solid rgba(165, 198, 230, .8);
    border-radius: 15px;
    background: rgba(255,255,255,.91);
    box-shadow: 0 18px 46px rgba(16, 61, 105, .16);
    backdrop-filter: blur(12px);
  }
  .kf-event-top { display: flex; justify-content: space-between; gap: 12px; margin-bottom: 11px; }
  .kf-event-type { color: #075ecf; font-size: 12px; font-weight: 800; letter-spacing: .1em; }
  .kf-event-time { color: #7a8ba0; font-size: 11px; }
  .kf-event-path { margin-bottom: 11px; color: #183854; font: 600 13px/1.45 ui-monospace, SFMono-Regular, Menlo, monospace; overflow-wrap: anywhere; }
  .kf-event-meta { display: flex; flex-wrap: wrap; gap: 6px; }
  .kf-event-meta span { padding: 4px 7px; border-radius: 6px; background: #edf5fc; color: #4a6278; font-size: 10px; font-weight: 650; }

  .kf-proof {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    margin: 0 0 clamp(76px, 10vw, 120px);
    border: 1px solid var(--kf-line);
    border-radius: 16px;
    background: #fff;
  }
  .kf-proof-item { padding: 18px 20px; border-right: 1px solid var(--kf-line); }
  .kf-proof-item:last-child { border-right: 0; }
  .kf-proof strong { display: block; margin-bottom: 2px; color: var(--kf-ink); font-size: 14px; }
  .kf-proof span { color: var(--kf-muted); font-size: 12px; }

  .kf-section { margin-bottom: clamp(80px, 11vw, 128px); }
  .kf-heading { max-width: 740px; margin-bottom: 42px; }
  .kf-heading h2 { margin-bottom: 15px; font-size: clamp(32px, 4vw, 48px); line-height: 1.08; }
  .kf-heading p { color: var(--kf-muted); font-size: 18px; }

  .kf-capabilities { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
  .kf-capability {
    min-height: 220px;
    padding: 30px;
    border: 1px solid var(--kf-line);
    border-radius: 18px;
    background: #fff;
    box-shadow: 0 10px 32px rgba(30, 76, 120, .05);
  }
  .kf-number { display: block; margin-bottom: 36px; color: #0b73dc; font: 700 12px/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .08em; }
  .kf-capability h3 { margin-bottom: 10px; font-size: 22px; }
  .kf-capability p { margin: 0; color: var(--kf-muted); font-size: 15px; }

  .kf-events { display: flex; flex-wrap: wrap; gap: 9px; margin-top: 24px; }
  .kf-events span { padding: 7px 11px; border: 1px solid #cfdef0; border-radius: 999px; background: #f8fbfe; color: #244b71; font: 700 11px/1 ui-monospace, SFMono-Regular, Menlo, monospace; letter-spacing: .06em; }

  .kf-flow {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    overflow: hidden;
    border: 1px solid #d5e5f5;
    border-radius: 20px;
    background: linear-gradient(135deg, #f8fbff, #edf7ff);
  }
  .kf-flow-step { position: relative; min-height: 190px; padding: 26px 20px; border-right: 1px solid #d5e5f5; }
  .kf-flow-step:last-child { border-right: 0; }
  .kf-flow-step::after { position: absolute; top: 31px; right: -7px; z-index: 1; width: 13px; height: 13px; border-top: 2px solid #8aafd3; border-right: 2px solid #8aafd3; background: #f3f9ff; content: ""; transform: rotate(45deg); }
  .kf-flow-step:last-child::after { display: none; }
  .kf-flow-index { display: block; margin-bottom: 30px; color: #5882ab; font-size: 11px; font-weight: 800; }
  .kf-flow h3 { margin-bottom: 9px; font-size: 17px; }
  .kf-flow p { margin: 0; color: var(--kf-muted); font-size: 13px; line-height: 1.55; }

  .kf-operations { display: grid; grid-template-columns: 1fr 1fr; gap: 18px; }
  .kf-operation { padding: clamp(28px, 5vw, 46px); border-radius: 22px; }
  .kf-operation h3 { margin-bottom: 14px; font-size: 28px; }
  .kf-operation p { color: #52657d; }
  .kf-operation-blue { border: 1px solid #cee2f7; background: linear-gradient(145deg, #edf7ff, #f8fcff); }
  .kf-operation-warm { border: 1px solid #e4e8ee; background: linear-gradient(145deg, #f8f9fb, #fff); }
  .kf-mini-list { display: grid; gap: 11px; margin-top: 26px; }
  .kf-mini-item { display: flex; gap: 12px; align-items: flex-start; color: #334d68; font-size: 14px; }
  .kf-mini-item::before { flex: 0 0 auto; width: 7px; height: 7px; margin-top: 8px; border-radius: 50%; background: var(--kf-cyan); content: ""; }
  .kf-endpoints { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 24px; }
  .kf-endpoints span { padding: 8px 10px; border-radius: 8px; background: #fff; color: #174b78; font: 650 12px/1 ui-monospace, SFMono-Regular, Menlo, monospace; box-shadow: inset 0 0 0 1px #d5e4f1; }

  .kf-install {
    display: grid;
    grid-template-columns: 1.05fr .95fr;
    gap: 30px;
    align-items: center;
    padding: clamp(34px, 6vw, 64px);
    border-radius: 24px;
    background: #112b49;
    color: #dceaff;
    box-shadow: 0 22px 54px rgba(12, 39, 69, .18);
  }
  .kf-install h2 { margin-bottom: 14px; color: #fff; font-size: clamp(32px, 4vw, 46px); line-height: 1.08; }
  .kf-install p { margin-bottom: 0; color: #bad0e8; }
  .kf-command { padding: 20px; border: 1px solid rgba(147, 189, 230, .25); border-radius: 14px; background: rgba(255,255,255,.07); }
  .kf-command-label { display: block; margin-bottom: 10px; color: #7fdff0; font-size: 11px; font-weight: 750; letter-spacing: .1em; text-transform: uppercase; }
  .kf-command code { color: #fff; font-size: 13px; white-space: normal; overflow-wrap: anywhere; }
  .kf-install-links { display: flex; gap: 16px; flex-wrap: wrap; margin-top: 18px; }
  .kf-install-links a { color: #fff !important; font-size: 13px; font-weight: 700; text-decoration-color: #5da7e9; text-underline-offset: 4px; }

  .kf-compat { display: grid; grid-template-columns: repeat(3, 1fr); gap: 14px; }
  .kf-compat-card { padding: 25px; border: 1px solid var(--kf-line); border-radius: 16px; background: #fff; }
  .kf-compat-label { display: block; margin-bottom: 18px; color: #0b6dcc; font-size: 11px; font-weight: 800; letter-spacing: .1em; text-transform: uppercase; }
  .kf-compat-card h3 { margin-bottom: 10px; font-size: 19px; }
  .kf-compat-card p { margin: 0; color: var(--kf-muted); font-size: 14px; }

  .kf-callout {
    position: relative;
    overflow: hidden;
    padding: clamp(38px, 7vw, 72px);
    border: 1px solid #cce1f5;
    border-radius: 24px;
    background: linear-gradient(135deg, #f8fcff, #eaf5ff);
    text-align: center;
  }
  .kf-callout::before { position: absolute; top: -110px; left: 50%; width: 520px; height: 260px; border-radius: 50%; background: rgba(25,185,209,.09); content: ""; transform: translateX(-50%); filter: blur(20px); }
  .kf-callout > * { position: relative; }
  .kf-callout h2 { margin-bottom: 14px; font-size: clamp(32px, 4vw, 46px); }
  .kf-callout p { max-width: 650px; margin: 0 auto 27px; color: var(--kf-muted); font-size: 17px; }
  .kf-callout .kf-actions { justify-content: center; }

  .kf-footnote { margin: 28px auto 8px; color: #6b7d91; font-size: 12px; text-align: center; }

  @media (max-width: 900px) {
    .kf-hero-grid, .kf-install { grid-template-columns: 1fr; }
    .kf-signal-stage { min-height: 300px; }
    .kf-proof { grid-template-columns: repeat(2, 1fr); }
    .kf-proof-item:nth-child(2) { border-right: 0; }
    .kf-proof-item:nth-child(-n+2) { border-bottom: 1px solid var(--kf-line); }
    .kf-flow { grid-template-columns: 1fr; }
    .kf-flow-step { min-height: auto; border-right: 0; border-bottom: 1px solid #d5e5f5; }
    .kf-flow-step:last-child { border-bottom: 0; }
    .kf-flow-step::after { display: none; }
    .kf-flow-index { margin-bottom: 14px; }
    .kf-compat { grid-template-columns: 1fr; }
  }

  @media (max-width: 620px) {
    .kf-hero { padding: 36px 22px; border-radius: 20px; }
    .kf-hero h1 { font-size: 41px; }
    .kf-signal-stage { min-height: 270px; }
    .kf-orbit { width: 250px; height: 250px; }
    .kf-capabilities, .kf-operations { grid-template-columns: 1fr; }
    .kf-capability { min-height: auto; }
    .kf-proof { grid-template-columns: 1fr; }
    .kf-proof-item { border-right: 0; border-bottom: 1px solid var(--kf-line) !important; }
    .kf-proof-item:last-child { border-bottom: 0 !important; }
    .kf-button { width: 100%; }
  }

  @media (prefers-reduced-motion: no-preference) {
    .kf-kicker::before { animation: kf-pulse 2.8s ease-out infinite; }
    .kf-event-card { animation: kf-float 5s ease-in-out infinite; }
    @keyframes kf-pulse { 0%, 50%, 100% { box-shadow: 0 0 0 6px rgba(32,183,137,.12); } 25% { box-shadow: 0 0 0 11px rgba(32,183,137,0); } }
    @keyframes kf-float { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-7px); } }
  }
</style>

<main class="kf-page">
  <div class="kf-shell">
    <section class="kf-hero" aria-labelledby="kf-title">
      <div class="kf-hero-grid">
        <div>
          <span class="kf-kicker">OWASP Incubator · Open source</span>
          <h1 id="kf-title">See every change. <span>Know the workload.</span></h1>
          <p class="kf-lede">KubeFIM turns Linux file activity and process execution into Kubernetes-aware security events—captured at the kernel, enriched on the node, and ready for the tools your team already operates.</p>
          <div class="kf-actions">
            <a class="kf-button kf-button-primary" href="https://github.com/OWASP/www-project-kubefim/tree/main/kubefim-src">Explore the source&nbsp; →</a>
            <a class="kf-button" href="https://github.com/OWASP/www-project-kubefim/blob/main/kubefim-src/README.md#install-on-kubernetes">Install KubeFIM</a>
          </div>
        </div>
        <div class="kf-signal-stage" aria-label="KubeFIM event visualization">
          <div class="kf-orbit" aria-hidden="true"></div>
          <img class="kf-logo" src="/www-project-kubefim/assets/images/kubefim-logo-v2.png" alt="OWASP KubeFIM logo">
          <div class="kf-event-card">
            <div class="kf-event-top"><span class="kf-event-type">EXEC</span><span class="kf-event-time">just now</span></div>
            <div class="kf-event-path">/usr/bin/curl</div>
            <div class="kf-event-meta"><span>payments</span><span>api-7c9f</span><span>uid 1000</span></div>
          </div>
        </div>
      </div>
    </section>

    <div class="kf-proof" aria-label="Project facts">
      <div class="kf-proof-item"><strong>One agent per node</strong><span>Kubernetes DaemonSet</span></div>
      <div class="kf-proof-item"><strong>AMD64 + ARM64</strong><span>Multi-architecture images</span></div>
      <div class="kf-proof-item"><strong>eBPF + Go</strong><span>Kernel signal, small agent</span></div>
      <div class="kf-proof-item"><strong>Apache 2.0</strong><span>Open source</span></div>
    </div>

    <section class="kf-section" aria-labelledby="kf-visibility">
      <div class="kf-heading">
        <span class="kf-section-kicker">Runtime context</span>
        <h2 id="kf-visibility">A kernel event is useful. A kernel event with workload identity is actionable.</h2>
        <p>KubeFIM observes activity below the container boundary, then connects it to the process and Kubernetes workload that caused it.</p>
        <div class="kf-events" aria-label="Observed operations"><span>OPEN</span><span>CREATE</span><span>DELETE</span><span>RENAME</span><span>CHMOD</span><span>CHOWN</span><span>EXECVE</span><span>EXECVEAT</span></div>
      </div>
      <div class="kf-capabilities">
        <article class="kf-capability"><span class="kf-number">01 / KERNEL</span><h3>Observe at the source</h3><p>eBPF tracepoint programs capture file operations and successful or failed process execution without modifying applications or injecting sidecars.</p></article>
        <article class="kf-capability"><span class="kf-number">02 / PROCESS</span><h3>Preserve execution context</h3><p>Events include PID, parent PID, user and group IDs, command, syscall result, namespaces, cgroup, and full container ID.</p></article>
        <article class="kf-capability"><span class="kf-number">03 / KUBERNETES</span><h3>Resolve the workload</h3><p>A node-scoped Pod cache adds namespace, Pod, container, image, image ID, and node metadata without an API request for every event.</p></article>
        <article class="kf-capability"><span class="kf-number">04 / POLICY</span><h3>Control noise safely</h3><p>Observe-only policy is the default. Explicit, reviewable rules can exclude known noise while counters keep filtered activity visible to operators.</p></article>
      </div>
    </section>

    <section class="kf-section" aria-labelledby="kf-pipeline">
      <div class="kf-heading">
        <span class="kf-section-kicker">Inside the agent</span>
        <h2 id="kf-pipeline">From syscall to structured event, on the node.</h2>
        <p>The data path stays compact: collect once, enrich locally, and emit formats that fit existing security operations.</p>
      </div>
      <div class="kf-flow">
        <article class="kf-flow-step"><span class="kf-flow-index">01</span><h3>Kernel probes</h3><p>Entry and exit tracepoints correlate operation, path, identity, and result.</p></article>
        <article class="kf-flow-step"><span class="kf-flow-index">02</span><h3>Ring buffer</h3><p>A stable binary layout carries events efficiently into user space.</p></article>
        <article class="kf-flow-step"><span class="kf-flow-index">03</span><h3>Go agent</h3><p>The collector decodes records and resolves process container identity.</p></article>
        <article class="kf-flow-step"><span class="kf-flow-index">04</span><h3>Pod cache</h3><p>A filtered list/watch cache adds Kubernetes metadata on each node.</p></article>
        <article class="kf-flow-step"><span class="kf-flow-index">05</span><h3>Policy + output</h3><p>Rules classify events before JSON and operational metrics are exposed.</p></article>
      </div>
    </section>

    <section class="kf-section" aria-labelledby="kf-operate">
      <div class="kf-heading">
        <span class="kf-section-kicker">Built to operate</span>
        <h2 id="kf-operate">Security data that does not require another platform.</h2>
      </div>
      <div class="kf-operations">
        <article class="kf-operation kf-operation-blue">
          <h3>Structured by default</h3>
          <p>One JSON object per line goes to standard output. Your Kubernetes log collector can route it to Elasticsearch, CloudWatch Logs, Loki, or another backend.</p>
          <div class="kf-mini-list"><span class="kf-mini-item">Workload and process identity in the same event</span><span class="kf-mini-item">No proprietary event transport</span><span class="kf-mini-item">Human-readable console mode for local work</span></div>
        </article>
        <article class="kf-operation kf-operation-warm">
          <h3>Observable in production</h3>
          <p>Alpha 3 ships fixed-cardinality Prometheus metrics, health checks, a ServiceMonitor overlay, and a provisioned Grafana dashboard.</p>
          <div class="kf-endpoints"><span>/metrics</span><span>/healthz</span><span>Grafana</span><span>Prometheus</span></div>
        </article>
      </div>
    </section>

    <section class="kf-section kf-install" aria-labelledby="kf-install-title">
      <div>
        <span class="kf-section-kicker" style="color:#7fdff0">Start on a Linux cluster</span>
        <h2 id="kf-install-title">Deploy one sensor on every node.</h2>
        <p>The initializer validates BTF and tracefs interfaces; it does not install kernel packages or download headers onto your nodes.</p>
        <div class="kf-install-links"><a href="https://github.com/OWASP/www-project-kubefim/blob/main/kubefim-src/README.md#requirements">Review requirements</a><a href="https://hub.docker.com/r/abhijitowasp/owasp-kubefim">View container images</a></div>
      </div>
      <div class="kf-command"><span class="kf-command-label">Deploy with Kustomize</span><code>kubectl apply -k kubefim-src/deployments/kubernetes</code></div>
    </section>

    <section class="kf-section" aria-labelledby="kf-compatibility">
      <div class="kf-heading">
        <span class="kf-section-kicker">Compatibility</span>
        <h2 id="kf-compatibility">Cloud-neutral by design. Kernel-dependent by nature.</h2>
        <p>KubeFIM uses Linux and Kubernetes interfaces rather than a provider API. Support depends on the node kernel and whether the cluster permits a privileged DaemonSet.</p>
      </div>
      <div class="kf-compat">
        <article class="kf-compat-card"><span class="kf-compat-label">Designed for</span><h3>Standard Linux workers</h3><p>Amazon EKS on EC2, GKE Standard, AKS, and self-managed Kubernetes clusters.</p></article>
        <article class="kf-compat-card"><span class="kf-compat-label">Verified today</span><h3>K3s on Ubuntu</h3><p>End-to-end on Ubuntu 24.04 ARM64, Linux 6.8, containerd, and K3s 1.36.1. Manifests API-validated on Kubernetes 1.34–1.36.</p></article>
        <article class="kf-compat-card"><span class="kf-compat-label">Current boundary</span><h3>Privileged access required</h3><p>Serverless and virtual nodes, plus managed modes that reject privileged DaemonSets, are not supported.</p></article>
      </div>
    </section>

    <section class="kf-section kf-callout" aria-labelledby="kf-community">
      <span class="kf-section-kicker">Build it with us</span>
      <h2 id="kf-community">Bring KubeFIM to more kernels and clusters.</h2>
      <p>Real-world compatibility reports and reproducible noise cases are especially valuable. Contributions across eBPF, Go, Kubernetes, testing, detection engineering, and documentation are welcome.</p>
      <div class="kf-actions"><a class="kf-button kf-button-primary" href="https://github.com/OWASP/www-project-kubefim/issues">Open an issue&nbsp; →</a><a class="kf-button" href="https://github.com/OWASP/www-project-kubefim/blob/main/CONTRIBUTING.md">Contribute</a><a class="kf-button" href="https://github.com/OWASP/www-project-kubefim/blob/main/SECURITY.md">Report a vulnerability</a></div>
    </section>

    <p class="kf-footnote">OWASP KubeFIM v0.1.0-alpha.3 · OWASP Incubator project · Apache License 2.0 · Observe-only by default</p>
  </div>
</main>
