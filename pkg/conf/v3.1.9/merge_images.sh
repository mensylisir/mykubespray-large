awk '
  BEGIN {
    images_header_printed = 0
  }
  /^[ ]{2}images:$/ {
    in_images_section = 1
    next
  }
  in_images_section {
    if (/^[ ]{2}- /) {
      if (!images_header_printed) {
        print "  images:"
        images_header_printed = 1
      }
      print
    } else {
      in_images_section = 0
    }
  }
' *.yaml
