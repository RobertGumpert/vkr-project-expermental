const inputId = "input";
const host = "http://127.0.0.1:60000";
const jsonViewModel = {
    keyword: "",
    name: "",
    owner: "",
    email: ""
};

window.onload = () => {
    document.getElementsByClassName("button-nearest-repositories-container")[0].addEventListener('click', () => {
        let nonValidInputValue = false;
        let inputData = {};
        Object.keys(jsonViewModel).forEach((field) => {
            let input = document.getElementById(
                [inputId, field].join('-')
            );
            if (input.value === "") {
                nonValidInputValue = true
            } else {
                input.style.backgroundColor = "white";
                inputData[field] = input.value;
            }
        });
        if (!nonValidInputValue) {
            const getNearestRepositories = [host, "get/nearest/repositories"].join("/");
            let request = new XMLHttpRequest();
            request.open(
                "POST",
                getNearestRepositories,
                false
            );
            request.setRequestHeader("Content-Type", "application/json");
            request.send(JSON.stringify(inputData));
            if (request.status !== 200) {
                setError();
            } else {
                const json = JSON.parse(request.response);
                console.log(json["task_state"]["endpoint"]);
                window.location = json["task_state"]["endpoint"];
            }
        } else {
            setError();
        }
    });
};

function setError() {
    Object.keys(jsonViewModel).forEach((field) => {
        let input = document.getElementById(
            [inputId, field].join('-')
        );
        input.value = "";
        input.placeholder = "Введите корректные данные.";
        input.style.backgroundColor = "#fce1e1";
    });
}
