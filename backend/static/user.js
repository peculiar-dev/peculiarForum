
init();

function init(){
    console.log("initialize");

    const img = document.getElementById("icon-img");
    if (img) {
      const timestamp = new Date().getTime();
      img.src = img.src.split('?')[0] + '?t=' + timestamp;
    }

    // JavaScript to handle pic button actions
    document.querySelectorAll('.pic-button').forEach(button => {
        button.addEventListener('click', function() {
            const picBoxId = this.getAttribute('data-pic-target');
            const picBox = document.getElementById(picBoxId);
            picBox.myfiles.click();
        });
    });

    // JavaScript to handle files button actions
    document.querySelectorAll('.files-button').forEach(button => {
        button.addEventListener('click', function() {
            const filesBoxId = this.getAttribute('data-files-target');
            const filesBox = document.getElementById(filesBoxId);
            filesBox.myfiles.click();
        });
    });
}   
 
// Used for creating a new FileList in a round-about way
function FileListItem(a) {
    a = [].slice.call(Array.isArray(a) ? a : arguments)
    for (var c, b = c = a.length, d = !0; b-- && d;) d = a[b] instanceof File
    if (!d) throw new TypeError("expected argument to FileList is File or array of File objects")
    for (b = (new ClipboardEvent("")).clipboardData || new DataTransfer; c--;) b.items.add(a[c])
    return b.files
}

async function autoSubmitIconForm(formID,event) {
    var form = document.getElementById(formID)
    if (form) {
        // Trigger the HTMX post
        // after resizing the image. 
            
        const file = event.target.files[0];
        if (!file) {
            console.log("no file found");
            return;
        }
        console.log("filename:" + file.name);

        /*resize here*/
        const maxWidth = 128;
        const maxHeight = 128;
        const result = [];
            
        for (const file of event.target.files) {
          const canvas = document.createElement('canvas');
          const ctx = canvas.getContext('2d');
          const img = await createImageBitmap(file);
              
          // calculate new size
          const ratio = Math.min(maxWidth / img.width, maxHeight / img.height);
          const width = img.width * ratio + .5 | 0;
          const height = img.height * ratio + .5 | 0;
          
          // resize the canvas to the new dimensions
          canvas.width = width;
          canvas.height = height;
          canvas.hidden = true;
          
          // scale & draw the image onto the canvas
          ctx.drawImage(img, 0, 0, width, height);

          //const pngData = canvas.toDataURL('image/png');

          //const pngData = await new Promise(rs => canvas.toBlob(rs, 1));
          
          // Get the binary (aka blob)
          const blob = await new Promise(rs => canvas.toBlob(rs,"image/png", 1));
          const resizedFile = new File([blob], file.name, file);
          //const resizedFile = new File([pngData], file.name, file);
          result.push(resizedFile);
        }
            
        const fileList = new FileListItem(result);
                  
              // temporary remove event listener since
              // assigning a new filelist to the input
              // will trigger a new change event...
              listener = form.onchange
              form.onchange = null
              event.target.files = fileList
              form.onchange = listener
        
        /* end resize */

        htmx.trigger(form, 'submit');
        console.log("filename:" + file.name);
            console.log("fired submit to autosubmit icon update form.");
    }
        
}

async function autoSubmitFilesForm(formID,event) {
    var form = document.getElementById(formID)
    if (form) {
        // Trigger the HTMX post
        // after resizing the image. 
            
        const file = event.target.files[0];
        if (!file) {
            console.log("no file found");
            return;
        }
        console.log("filename:" + file.name);
        htmx.trigger(form, 'submit');
        console.log("filename:" + file.name);
            console.log("fired submit to autosubmit Files update form.");
    }
        
}

document.addEventListener('htmx:afterSettle',function(evt){
    init();            
});