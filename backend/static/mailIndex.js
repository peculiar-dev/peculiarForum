    var current_comment = "";
    var current_root = window.location.pathname.split('/')[2];
   
    init();
    function autoSubmitForm(formID) {
        //document.getElementById(formID).submit();
        var form = document.getElementById(formID)
        if (form) {
            // Trigger the HTMX post
            htmx.trigger(form, 'submit');
        }
    }

    function init(){
        console.log("initialize");
        console.log("root:",current_root);
        //document.getElementById("form_root").value = current_root;
        var elements = document.getElementsByName("root");
        for(var i = 0; i < elements.length; i++) {
            console.log("found");
            elements[i].value = current_root;
        }



    // JavaScript to handle collapsible actions
    document.querySelectorAll('.collapse-button').forEach(button => {
        button.addEventListener('click', function() {
            const targetId = this.getAttribute('data-target');
            const content = document.getElementById(targetId);
            if (content.style.display === 'block') {
                content.style.display = 'none';
                this.innerHTML = '+';
            } else {
                content.style.display = 'block';
                this.innerHTML = '-';
            }
        });
    });

    // JavaScript to handle reply button actions
    document.querySelectorAll('.reply-button').forEach(button => {
        button.addEventListener('click', function() {
            const replyBoxId = this.getAttribute('data-reply-target');
            const replyBox = document.getElementById(replyBoxId);
            if (replyBox.style.display === 'block') {
                replyBox.style.display = 'none';
            } else {
                replyBox.style.display = 'block';
            }
        });
    });

    // JavaScript to handle edit button actions
    document.querySelectorAll('.edit-button').forEach(button => {
        button.addEventListener('click', function() {
            const editBoxId = this.getAttribute('data-edit-target');
            const editBox = document.getElementById(editBoxId);
            if (editBox.style.display === 'block') {
                editBox.style.display = 'none';
            } else {
                editBox.style.display = 'block';
            }
        });
    });


    // JavaScript to handle pic button actions
    document.querySelectorAll('.pic-button').forEach(button => {
        button.addEventListener('click', function() {
            const picBoxId = this.getAttribute('data-pic-target');
            const picBox = document.getElementById(picBoxId);
            picBox.myfiles.click();
        });
    });

    // Handle sending (example)
    document.querySelectorAll('.send-button').forEach(button => {
        button.addEventListener('click', function(ccomment) {
            current_comment = ccomment;
        });
    });

    // make post box hidden on each HTMX refresh
    const postBox = document.getElementById("post-box");
    postBox.style.display = 'none';

    }

        // JavaScript to handle post button actions
        document.querySelectorAll('.post-button').forEach(button => {
            console.log("initializing post button")
            const postBox = document.getElementById("post-box");
            postBox.style.display = 'none';
            button.addEventListener('click', function() {
                if (postBox.style.display === 'block') {
                    postBox.style.display = 'none';
                } else {
                    postBox.style.display = 'block';
                }
            });
        });