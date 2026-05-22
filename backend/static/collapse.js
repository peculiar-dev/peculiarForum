var current_comment = "";
var current_root = window.location.pathname.split('/')[2];

console.log(window.location.pathname.split('/')[3])

if (window.location.pathname.split('/')[3]) {
    current_comment = "reply-box-" + window.location.pathname.split('/')[3];
}

//init();

window.addEventListener('load', function () {
    init();
});

//function to convert creation timestamp to a time-since human readable text
function timeSince(date) {

    var seconds = Math.floor(((new Date() / 1000) - date));
    var interval = Math.floor(seconds / 31536000);

    if (interval < 0) {
        return " Sticky";
    }

    //console.log("interval:"+interval);

    if (interval == 1) {
        return interval + " year ago";
    }
    if (interval > 1) {
        return interval + " years ago";
    }
    interval = Math.floor(seconds / 2592000);
    if (interval == 1) {
        return interval + " month ago";
    }
    if (interval > 1) {
        return interval + " months ago";
    }
    interval = Math.floor(seconds / 86400);
    if (interval == 1) {
        return interval + " day ago";
    }
    if (interval > 1) {
        return interval + " days ago";
    }
    interval = Math.floor(seconds / 3600);
    if (interval == 1) {
        return interval + " hour ago";
    }
    if (interval > 1) {
        return interval + " hours ago";
    }
    interval = Math.floor(seconds / 60);
    if (interval == 1) {
        return interval + " minute ago";
    }
    if (interval > 1) {
        return interval + " minutes ago";
    }
    return Math.floor(seconds) + " seconds ago";
}

// Used for creating a new FileList in a round-about way
function FileListItem(a) {
    a = [].slice.call(Array.isArray(a) ? a : arguments)
    for (var c, b = c = a.length, d = !0; b-- && d;) d = a[b] instanceof File
    if (!d) throw new TypeError("expected argument to FileList is File or array of File objects")
    for (b = (new ClipboardEvent("")).clipboardData || new DataTransfer; c--;) b.items.add(a[c])
    return b.files
}

async function autoSubmitForm(formID, event) {
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
        const maxWidth = 800;
        const maxHeight = 800;
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

            // Get the binary (aka blob)
            const blob = await new Promise(rs => canvas.toBlob(rs, 1));
            const resizedFile = new File([blob], file.name, file);
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
        console.log("fired submit to autosubmit picture form.");
    }

}

// post-button click function
function postClick(e) {
    const postBox = document.getElementById("post-box");
    if (postBox.style.display === 'block') {
        postBox.style.display = 'none';
    } else {
        postBox.style.display = 'block';
    }
};

//collaps-button click function
function collapseClick(e) {
    const targetId = e.currentTarget.getAttribute('data-target');
    const content = document.getElementById(targetId);
    console.log("target display:" + content.style.display);
    if (content.style.display === 'block' || content.style.display === '') {
        content.style.display = 'none';
        e.currentTarget.innerHTML = '+';
    } else {
        content.style.display = 'block';
        e.currentTarget.innerHTML = '-';
    }
};

// reply button click function
function replyClick(e) {
    const replyBoxId = e.currentTarget.getAttribute('data-reply-target');
    console.log("Reply Box ID:", replyBoxId);
    const replyBox = document.getElementById(replyBoxId);
    console.log("replyBox:", replyBox);
    if (replyBox.style.display === 'block') {
        replyBox.style.display = 'none';
    } else {
        replyBox.style.display = 'block';
    }
};

// edit button click function
function editClick(e) {
    const editBoxId = e.currentTarget.getAttribute('data-edit-target');
    const editBox = document.getElementById(editBoxId);
    if (editBox.style.display === 'block') {
        editBox.style.display = 'none';
    } else {
        editBox.style.display = 'block';
    }
};

// pic button click function
function picClick(e) {
    const picBoxId = e.currentTarget.getAttribute('data-pic-target');
    const picBox = document.getElementById(picBoxId);
    picBox.myfiles.click();
};

// function to handle img-click
function imgClick(e) {
    console.log("image clicked");
    const img = e.currentTarget;
    if (img.width != img.naturalWidth || img.height != img.naturalHeight) {
        console.log("resizing image to natural dimensions");
        img.style.width = img.naturalWidth + "px";
        img.style.height = img.naturalHeight + "px";
    } else {
        console.log("resizing image to default dimensions");
        const rootStyles = getComputedStyle(document.documentElement);
        const imgDefaultWidth = rootStyles.getPropertyValue('--img-default-width').trim();
        console.log("imgDefaultWidth:", imgDefaultWidth);
        img.style.width = imgDefaultWidth;
        img.style.height = "auto";
    }
}

function init() {
    console.log("initialize");
    console.log("root:", current_root);
    console.log("current:", current_comment);
    //document.getElementById("form_root").value = current_root;
    var elements = document.getElementsByName("root");
    for (var i = 0; i < elements.length; i++) {
        console.log("found");
        elements[i].value = current_root;
    }

    // JavaScript to handle post button actions
    document.querySelectorAll('.post-button').forEach(button => {
        button.addEventListener('click', postClick)
    });

    // JavaScript to handle collapsible actions
    document.querySelectorAll('.collapse-button').forEach(button => {
        button.addEventListener('click', collapseClick)
    });

    // JavaScript to handle reply button actions
    document.querySelectorAll('.reply-button').forEach(button => {
        button.addEventListener('click', replyClick);
    });

    // JavaScript to handle edit button actions
    document.querySelectorAll('.edit-button').forEach(button => {
        button.addEventListener('click', editClick);
    });

    // JavaScript to handle pic button actions
    document.querySelectorAll('.pic-button').forEach(button => {
        button.addEventListener('click', picClick);
    });

    // Handle sending (example)
    document.querySelectorAll('.send-button').forEach(button => {
        button.addEventListener('click', function (ccomment) {
            current_comment = ccomment;
        });
    });

    // JavaScript to handle img click actions
    console.log("adding click event to images");
    document.querySelectorAll('.post-pic').forEach(img => {
        img.addEventListener('click', imgClick)
    });

    console.log("current comment before if:" + current_comment);
    if (current_comment) {
        recent = document.getElementById(current_comment).parentElement.children[1]
        console.log("recent classname:" + recent.className)
        console.log("recent tagname:" + recent.tagName)
        if (recent) {
            recent.focus();
            currentItem = recent;
            while (currentItem && currentItem.tagName !== 'BODY') {
                if (currentItem.className === "collapsible-content") {
                    currentItem.style.display = 'block';
                    console.log("Current Item ID:" + currentItem.id)
                    console.log("Current Item Style.display:" + currentItem.style.display)
                }
                currentItem = currentItem.parentElement;
            }

        }
    }

}

document.addEventListener('htmx:afterSettle', function (evt) {
    if(evt.detail?.elt?.id === 'chat-unread-count') return;
    init();

});