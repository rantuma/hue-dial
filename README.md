# Hue Tap Dial Switch

The Philips Hue Tap Dial Switch has four buttons and a rotating dial. In the official Hue app, each button can only be assigned to a Room, Zone, or the whole house — activating a scene for that group. The dial then adjusts the brightness of whatever group was last activated. There is no way to map a button to an individual lamp, and the dial always targets the entire group rather than a single bulb.

I expected to use the four buttons to each select one specific lamp in my home and then use the dial to fine-tune the brightness of just that lamp — a simple "one button = one light, dial = brightness" workflow. Since the official app doesn't support this, I wrote this Go service.

This service reacts to Philips Hue Tap Dial Switch events and controls lights via the Hue Bridge API (CLIP v2).

- **Button short press** — selects the mapped lamp and triggers its identify animation.
- **Button long press** — turns the mapped lamp off.
- **Dial rotate** — adjusts brightness of the selected lamp. Clockwise increases, counter-clockwise decreases.
- **Dial inverted** — if the dial is mounted upside down, enable `invert_dial` in the config to flip the rotation direction. The setup wizard prompts for this automatically.

## Quick start (Docker)

No repo clone needed. The image is published on GitHub Container Registry as [`ghcr.io/rantuma/hue-dial`](https://github.com/rantuma/hue-dial/pkgs/container/hue-dial).

**1. Pull the image**

```bash
docker pull ghcr.io/rantuma/hue-dial
```

**2. Run the setup wizard (first time only)**

```bash
mkdir -p data
docker run -it --rm --network host -v "$(pwd)/data:/data" ghcr.io/rantuma/hue-dial
```

The wizard auto-discovers the Hue Bridge, asks you to press the link button, and lets you map each Tap Dial button to a lamp. Config is saved to `./data/config.json`. Once done, stop the container (`Ctrl-C`).

**3. Start the service**

```bash
docker run -d \
  --name hue-dial \
  --network host \
  --restart unless-stopped \
  -v "$(pwd)/data:/data" \
  ghcr.io/rantuma/hue-dial
```

`--network host` is required to reach the Hue Bridge on the local network.

**Re-run the setup wizard**

```bash
docker run -it --rm --network host -v "$(pwd)/data:/data" ghcr.io/rantuma/hue-dial --setup
```

**Config path resolution**

When `CONFIG_PATH` is not set, the service tries each location in order and uses the first one whose parent directory is writable:

| Priority | Path |
|----------|------|
| 1 | `/data/config.json` — default inside Docker |
| 2 | `$XDG_CONFIG_HOME/hue-dial/config.json` (Linux) |
| 2 | `~/Library/Application Support/hue-dial/config.json` (macOS) |
| 2 | `%AppData%\hue-dial\config.json` (Windows) |
| 3 | `./config.json` — current working directory |

**Override the config path**

```bash
docker run -d \
  --name hue-dial \
  --network host \
  --restart unless-stopped \
  -e CONFIG_PATH=/data/my-config.json \
  -v "$(pwd)/data:/data" \
  ghcr.io/rantuma/hue-dial
```

## Updating to the latest version

**1. Pull the latest image**

```bash
docker pull ghcr.io/rantuma/hue-dial
```

**2. Stop and remove the running container**

```bash
docker stop hue-dial && docker rm hue-dial
```

**3. Start the service again**

```bash
docker run -d \
  --name hue-dial \
  --network host \
  --restart unless-stopped \
  -v "$(pwd)/data:/data" \
  ghcr.io/rantuma/hue-dial
```

Your `./data/config.json` is preserved — no reconfiguration needed.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Licenses

This project is licensed under the [MIT License](LICENSE).

Third-party dependency licenses are collected in the [`licenses/`](licenses/) directory. 

Thanks to everyone who builds and maintains open-source software!
