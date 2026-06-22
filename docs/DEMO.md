# QuantumField real demo capture

The repository does not claim a hosted demo or use generated screenshots as proof of a working deployment. Demo artifacts must be captured from the running Compose stack.

## Before recording

```bash
cp .env.example .env
make dev
```

In a second terminal:

```bash
make seed
```

Wait until `http://localhost:8080/health` returns `{"status":"ok"}`, then open `http://localhost:3000`.

## Suggested 60–90 second recording

1. Open the login page and sign in with the seeded demo account.
2. Open **Assets** and add a public domain such as `example.com`.
3. Leave **Queue the first scan immediately** enabled.
4. Open **Scan jobs** and show the job move from queued/running to completed.
5. Open the asset detail page and briefly show:
   - certificate identity, issuer, validity, and fingerprint;
   - chain and hostname validation;
   - TLS version and cipher suite;
   - findings, evidence, and remediation;
   - risk and crypto-agility scores.
6. Finish on the dashboard or reports page.

## Required real screenshots

Capture these files directly from the running application:

```text
docs/screenshots/login.png
docs/screenshots/dashboard.png
docs/screenshots/add-asset.png
docs/screenshots/scan-running.png
docs/screenshots/asset-detail.png
docs/screenshots/certificate-details.png
docs/screenshots/findings.png
docs/screenshots/report-export.png
```

Do not commit empty placeholders or generated approximations. After capturing them, add only the strongest four to the main README and keep the rest as supporting evidence.

## Video output

Save the final 60–90 second recording as `docs/demo/quantumfield-demo.mp4`, or upload it to a stable public host and add the link under a `## Demo` heading in the README.

If a GIF is needed, create a short, optimized excerpt rather than converting the entire recording:

```bash
ffmpeg -i docs/demo/quantumfield-demo.mp4 \
  -vf "fps=12,scale=1280:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
  docs/demo/quantumfield-demo.gif
```

## Recording guidance

- Use a real local Docker Compose run.
- Do not label the crypto-agility assessment as implemented post-quantum cryptography.
- Keep secrets, local paths, and unrelated browser tabs out of frame.
- Export as an optimized GIF or MP4 and link it from the README only after the artifact is committed or hosted.
- Keep the browser at a readable zoom level and avoid cuts that conceal scan state changes.
