# Go Hide

![Gopher with hidden message](https://github.com/JulianKnodt/gohide/blob/master/out.png)

Hide text in images.

This hides text in images by taking in an image and using the alpha channels as a medium to hold text.

It can store as many bytes as there are pixels in the image.

## Usage

```
./gohide -f <INPUT.PNG/.JPG> -msg "YOUR TEXT HERE" > out.txt

# Copy key from out .txt

./gohide -f out.png -key "COPY OF KEY"
```

Sample

Key for gopher above: "1041 210 421"
