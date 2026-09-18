from PIL import Image
import numpy as np

src = r"e:\wenlv\wennv\wenlv\public\images\culture-scroll\era-scroll-lineart-v1.png"
out = r"e:\wenlv\wennv\wenlv\public\images\culture-scroll\era-scroll-lineart-transparent-v1.png"

img = Image.open(src).convert("RGBA")
arr = np.asarray(img).astype(np.float32)
rgb = arr[:, :, :3]
lum = 0.299 * rgb[:, :, 0] + 0.587 * rgb[:, :, 1] + 0.114 * rgb[:, :, 2]
alpha = np.clip((220.0 - lum) / (220.0 - 40.0), 0.0, 1.0)
out_arr = np.zeros_like(arr)
out_arr[:, :, 0:3] = 26.0
out_arr[:, :, 3] = alpha * 255.0
out_arr[:, :, 3] = np.where(out_arr[:, :, 3] < 8, 0, out_arr[:, :, 3])

Image.fromarray(out_arr.astype(np.uint8), "RGBA").save(out, "PNG")
print("wrote", out)
print("size", Image.open(out).size)
print("nonzero_alpha", int(np.count_nonzero(out_arr[:, :, 3] > 0)))
