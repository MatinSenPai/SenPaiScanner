"""Regenerate every icon from logo/logo.png (white ink + one red accent on black).

    python gen_icons.py            # needs Pillow

Outputs: desktop app icon (png + ico), the GUI header mark, and the Android launcher set.
"""
import os
from PIL import Image, ImageDraw

ROOT = os.path.dirname(os.path.abspath(__file__))
RES = os.path.join(ROOT, "android", "app", "src", "main", "res")
BLACK = (0, 0, 0, 255)


def load_art(path):
    """Crop the logo to its ink, add a little margin, and return it as a black square."""
    img = Image.open(path).convert("RGB")
    ink = img.convert("L").point(lambda v: 255 if v > 24 else 0)
    img = img.crop(ink.getbbox())
    side = int(max(img.size) * 1.08)
    sq = Image.new("RGBA", (side, side), BLACK)
    sq.paste(img, ((side - img.width) // 2, (side - img.height) // 2))
    return sq


def fit(art, size, scale=1.0):
    """Art scaled to `scale` of a size x size black canvas, centred."""
    inner = max(1, int(size * scale))
    out = Image.new("RGBA", (size, size), BLACK)
    a = art.resize((inner, inner), Image.Resampling.LANCZOS)
    out.paste(a, ((size - inner) // 2, (size - inner) // 2))
    return out


def circle(img):
    mask = Image.new("L", img.size, 0)
    ImageDraw.Draw(mask).ellipse([(0, 0), (img.width - 1, img.height - 1)], fill=255)
    out = Image.new("RGBA", img.size, (0, 0, 0, 0))
    out.paste(img, (0, 0), mask)
    return out


def save(img, *parts, **kw):
    path = os.path.join(ROOT, *parts)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    img.save(path, **kw)


def main():
    art = load_art(os.path.join(ROOT, "logo", "logo.png"))
    droid = ("android", "app", "src", "main", "res")

    save(fit(art, 1024), "logo", "logo.png")
    save(fit(art, 1024, 0.9), "desktop", "build", "appicon.png")
    save(fit(art, 192, 0.96), "desktop", "frontend", "dist", "logo.png")
    save(fit(art, 256, 0.96), "desktop", "build", "windows", "icon.ico",
         sizes=[(16, 16), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)])

    # Android: in-app logo, adaptive foreground (art inside the 66% safe zone), legacy square/round launchers.
    save(fit(art, 1024), *droid, "drawable-nodpi", "senpai_logo.png")
    save(fit(art, 432, 0.62), *droid, "drawable-nodpi", "ic_launcher_foreground_adaptive.png")
    save(fit(art, 460), *droid, "drawable", "ic_launcher_foreground_raw.png")
    for density, size in {"mdpi": 48, "hdpi": 72, "xhdpi": 96, "xxhdpi": 144, "xxxhdpi": 192}.items():
        square = fit(art, size, 0.9)
        save(square, *droid, f"mipmap-{density}", "ic_launcher.png")
        save(circle(square), *droid, f"mipmap-{density}", "ic_launcher_round.png")

    with open(os.path.join(RES, "values", "colors.xml"), "w", encoding="utf-8", newline="\n") as f:
        f.write('<?xml version="1.0" encoding="utf-8"?>\n<resources>\n'
                '    <color name="ic_launcher_background">#000000</color>\n</resources>\n')
    print("icons regenerated from logo/logo.png")


if __name__ == "__main__":
    main()
