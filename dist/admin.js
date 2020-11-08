"use strict";
function deletePost() {
    let req = new XMLHttpRequest();
    req.addEventListener("load", function (event) {
        alert("Post deleted!");
        window.location.href = '/';
    });
    req.open("DELETE", "/api/v1" + window.location.pathname);
    req.send();
}
function editPost() {
    let postContent = document.querySelector("#post-content").textContent;
    let template = document.querySelector("#edit-form-template").innerHTML;
    let renderFunction = doT.template(template);
    let rendered = renderFunction();
    document.querySelector("#admin-container").innerHTML = rendered;
    let contentInput = document.querySelector("#content-input");
    contentInput.value = postContent;
    let saveChangesBtn = document.querySelector("#save-changes-btn");
    saveChangesBtn.addEventListener("click", saveChanges);
}
function saveChanges() {
    let titleInput = document.querySelector("#title-input");
    let textInput = document.querySelector("#content-input");
    let publicInput = document.querySelector("#public-input");
    let req = new XMLHttpRequest();
    req.addEventListener("load", function (event) {
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
    let editBtn = document.querySelector("#edit-btn");
    let deleteBtn = document.querySelector("#delete-btn");
    editBtn.addEventListener("click", editPost);
    deleteBtn.addEventListener("click", deletePost);
}
attachAdminButtonListeners();
