function deletePost() {
    let req = new XMLHttpRequest();

    req.addEventListener("load", function(event) {
        alert("Post deleted!");
        window.location.href = '/';
    });

    req.open("DELETE", "/api/v1" + window.location.pathname);
    req.send();
}

function editPost() {
    // Get the current post content
    let postContent = document.querySelector("#post-content").textContent;

    // Get the template from the DOM.
    let template = document.querySelector("#edit-form-template").innerHTML;

    // Create a render function for the template with doT.template.
    let renderFunction = doT.template(template);

    // Use the render function to render the data.
    let rendered = renderFunction();

    // Insert the result into the DOM (inside the <div> with the ID log.
    document.querySelector("#admin-container").innerHTML = rendered;

    // Set content input to the current post content
    let contentInput = <HTMLTextAreaElement> document.querySelector("#content-input");
    contentInput.value = postContent;

    // Add event listener to the newly-added button
    let saveChangesBtn = <HTMLInputElement> document.querySelector("#save-changes-btn");
    saveChangesBtn.addEventListener("click", saveChanges);
}

function saveChanges() {
    let titleInput : HTMLInputElement = document.querySelector("#title-input");
    let textInput : HTMLInputElement = document.querySelector("#content-input");
    let publicInput : HTMLInputElement = document.querySelector("#public-input");

    let req = new XMLHttpRequest();

    req.addEventListener("load", function(event) {
        window.location.reload();
    });

    req.open("PUT", "/api/v1" + window.location.pathname);
    req.send(JSON.stringify({
        title: titleInput.value,
        text: textInput.value,
        public: publicInput.checked,
    }));
}

function attachAdminButtonListeners() {
    let editBtn = <HTMLInputElement> document.querySelector("#edit-btn");
    let deleteBtn = <HTMLInputElement> document.querySelector("#delete-btn");
    
    editBtn.addEventListener("click", editPost);
    deleteBtn.addEventListener("click", deletePost);
}

attachAdminButtonListeners();