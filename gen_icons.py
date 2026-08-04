import os
import sys
from PIL import Image, ImageDraw

# Pillow compatibility: _resample was added in Pillow 9.0.0
try:
    _resample = Image.Resampling.LANCZOS
except AttributeError:
    _resample = Image.LANCZOS

def create_rounded_rect(size, radius, fill_color):
    """Create a rounded rectangle image"""
    img = Image.new('RGBA', size, (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    draw.rounded_rectangle([(0, 0), size], radius, fill=fill_color)
    return img

def create_circle(size, fill_color):
    """Create a circular image"""
    img = Image.new('RGBA', size, (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    draw.ellipse([(0, 0), size], fill=fill_color)
    return img

def generate_icons(project_dir):
    res_dir = os.path.join(project_dir, 'android', 'app', 'src', 'main', 'res')
    foreground_raw_path = os.path.join(res_dir, 'drawable', 'ic_launcher_foreground_raw.png')
    
    if not os.path.exists(foreground_raw_path):
        print(f"File not found: {foreground_raw_path}")
        return

    # Background color is #F6821F
    bg_color = (246, 130, 31, 255)
    
    # Load foreground and crop transparency
    fg_img = Image.open(foreground_raw_path).convert("RGBA")
    bbox = fg_img.getbbox()
    if bbox:
        fg_img = fg_img.crop(bbox)
    
    densities = {
        'mdpi': 48,
        'hdpi': 72,
        'xhdpi': 96,
        'xxhdpi': 144,
        'xxxhdpi': 192
    }
    
    for density, size in densities.items():
        mipmap_dir = os.path.join(res_dir, f'mipmap-{density}')
        os.makedirs(mipmap_dir, exist_ok=True)
        
        # Calculate foreground size (about 75% of total size to allow padding)
        # Standard icon is 48dp, safe zone is inner 30dp or so
        fg_size = int(size * 0.75)
        fg_resized = fg_img.resize((fg_size, fg_size), _resample)
        
        # Center coordinate
        offset = ((size - fg_size) // 2, (size - fg_size) // 2)
        
        # 1. Square / Rounded Rect icon (ic_launcher.png)
        # Android 7 and below standard is rounded rect or square. Let's do rounded rect (radius = size * 0.08)
        radius = int(size * 0.08)
        square_bg = create_rounded_rect((size, size), radius, bg_color)
        square_bg.paste(fg_resized, offset, fg_resized)
        square_bg.save(os.path.join(mipmap_dir, 'ic_launcher.png'))
        
        # 2. Round icon (ic_launcher_round.png)
        round_bg = create_circle((size, size), bg_color)
        round_bg.paste(fg_resized, offset, fg_resized)
        round_bg.save(os.path.join(mipmap_dir, 'ic_launcher_round.png'))
        
        # 3. Save a properly scaled foreground PNG to be used directly in adaptive icon
        # For adaptive icons, the base size is 108dp. But it's vector-based.
        # If we use PNG, it should be 108x108.
        # We'll generate a 432x432 (xxxhdpi equivalent for 108dp) transparent foreground
        if density == 'xxxhdpi':
            adaptive_size = 432
            adaptive_fg_size = int(432 * 0.8) # Foreground takes ~80% to fill space well without clipping too much
            adaptive_fg_resized = fg_img.resize((adaptive_fg_size, adaptive_fg_size), _resample)
            adaptive_fg = Image.new('RGBA', (adaptive_size, adaptive_size), (0, 0, 0, 0))
            adaptive_offset = ((adaptive_size - adaptive_fg_size) // 2, (adaptive_size - adaptive_fg_size) // 2)
            adaptive_fg.paste(adaptive_fg_resized, adaptive_offset, adaptive_fg_resized)
            
            # Save it to drawable-nodpi so we don't need the inset xml
            nodpi_dir = os.path.join(res_dir, 'drawable-nodpi')
            os.makedirs(nodpi_dir, exist_ok=True)
            adaptive_fg.save(os.path.join(nodpi_dir, 'ic_launcher_foreground_adaptive.png'))

    print(f"Generated icons for {project_dir}")

if __name__ == '__main__':
    generate_icons(r'C:\Users\user\Desktop\app\goose\New folder\SenPaiScanner-main')
