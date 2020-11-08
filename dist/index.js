"use strict";
let submitBtn = document.querySelector("#submit-btn");
let getPostsBtn = document.querySelector("#get-posts-btn");
function getRecentPosts(page = 1) {
    let req = new XMLHttpRequest();
    req.addEventListener("load", function (event) {
        let data = JSON.parse(req.responseText);
        console.log(data);
        data.results = data.results.slice(0, 5);
        updateViewMain(data);
    });
    req.open("GET", "/api/v1/posts?page=" + page);
    req.send();
}
function updateViewMain(data) {
    let template = document.querySelector("#recent-posts-template").innerHTML;
    let renderFunction = doT.template(template);
    let rendered = renderFunction(data);
    document.querySelector("#recent-posts").innerHTML = rendered;
}
function createPost() {
    let titleInput = document.querySelector("#title-input");
    let textInput = document.querySelector("#content-input");
    let publicInput = document.querySelector("#public-input");
    let req = new XMLHttpRequest();
    req.addEventListener("load", function () {
        let data = JSON.parse(req.responseText);
        window.location.href = "/posts/" + data.postContent.writeId;
    });
    req.open("POST", "/api/v1/posts");
    req.send(JSON.stringify({
        title: titleInput.value,
        text: textInput.value,
        public: publicInput.checked,
    }));
}
function attachMainButtonlisteners() {
    submitBtn.addEventListener("click", createPost);
    getPostsBtn.addEventListener("click", function () {
        window.location.href = '/posts';
    });
}
getRecentPosts();
attachMainButtonlisteners();
