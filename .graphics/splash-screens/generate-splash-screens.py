# Script on Python-fu for GIMP: generation splash-screens .png for any devices.

# An image must be stored in splash.xcf file (should not be open in advance), at least 2732 x 2732 pixels in size, containing layers: 
# "background" - a background image without rounded corners, exported from Iconfu to .svg, 
#    then rasterized in GIMP in the size of 2732 x 2732 pixels
# "logo" - a layer containing an application icon on a transparent background, 
#    also imported from Iconfu .svg to 2732 x 2732 pixels - the size of the icon should not exceed 50%.
# All two layers must be visible.

# HOW TO RUN:
# 1. This file: Ctrl-A, Ctrl-C
# 2. GIMP: Filters -> Python-Fu -> Console: Ctlr-V

# WARNING! BEFORE RUNNING, CHECK PATH:
path = "e:\\GitHub\\KornToDoList\\.graphics\\splash-screens\\"

# Sub-routine: reduces the image to the specified size and saves in PNG format
def make_png(width, height):

    # Work with current image
    image = pdb.gimp_xcf_load(0, path+"splash.xcf", path+"splash.xcf")
    # Bring image to the maximum size
    max_size = max(width,height)
    pdb.gimp_image_scale(image, max_size, max_size)
    # Resize logo
    layer = image.layers[0]
    min_size = min(width,height)
    layer.scale( int(layer.width*min_size/max_size), int(layer.height*min_size/max_size) )
    layer.set_offsets( (image.width-layer.width)//2, (image.height-layer.height)//2 )
    # Crop image
    pdb.gimp_image_crop(image, width, height, (max_size-width)//2, (max_size-height)//2)
    # Save PNG-file
    file_name = path+`width`+"x"+`height`+".png"
    flat_image = pdb.gimp_image_duplicate(image)
    flat_layer = pdb.gimp_image_merge_visible_layers(flat_image, CLIP_TO_IMAGE)
    pdb.file_png_save_defaults(flat_image, flat_layer, file_name, file_name)
    pdb.gimp_image_delete(flat_image)
    # Close original image without saving
    pdb.gimp_image_delete(image)

# Generate PNG-files
make_png(2048,2732)
make_png(1668,2388)
make_png(1668,2224)
make_png(1536,2048)
make_png(1242,2688)
make_png(1242,2208)
make_png(1125,2436)
make_png(828,1792)
make_png(750,1334)
make_png(640,1136)

make_png(2732,2048)
make_png(2388,1668)
make_png(2224,1668)
make_png(2048,1536)
make_png(2688,1242)
make_png(2208,1242)
make_png(2436,1125)
make_png(1792,828)
make_png(1334,750)
make_png(1136,640)

# End of script ;)

