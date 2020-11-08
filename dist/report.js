"use strict";
function showReportForm() {
    let template = document.querySelector("#report-form-template").innerHTML;
    let renderFunction = doT.template(template);
    let rendered = renderFunction();
    document.querySelector("#report-form").innerHTML = rendered;
    let reportSubmitBtn = document.querySelector("#report-submit-btn");
    reportSubmitBtn.addEventListener("click", submitReportForm);
}
function submitReportForm() {
    let reason = document.querySelector('input[name="report-reason"]:checked');
    createReport(reason.value);
}
function createReport(reason) {
    let req = new XMLHttpRequest();
    req.addEventListener("load", function (event) {
        alert("Post reported!");
        window.location.href = '/';
    });
    req.open("POST", "/api/v1" + window.location.pathname + "/reports");
    req.send(JSON.stringify({
        reason: reason,
    }));
}
let reportBtn = document.querySelector("#report-btn");
reportBtn.addEventListener("click", showReportForm);
