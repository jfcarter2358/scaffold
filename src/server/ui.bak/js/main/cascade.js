// import workflow.js
// import material.js
// import theme.js
// import modal.js
// import user_menu.js
// import deploy_status.js
// import data.js
// import run_state.js
// import input.js
// import accordion.js

stateIntervalMilliSeconds = 500
workflowIntervalMilliSeconds = 50

var tasks = {}
var states = []
var datastore = {}
var link_data = []
var node_data = []
var structure = {}
var elements = []
var positions = {}

right_panel_width = 500

var workflow

pin_colors = {
    "Success": "#A3BE8C",
    "Error": "#BF616A",
    "Always": "#5E81AC",
}

var state_colors = {
    "not_started": "scaffold-charcoal",
    "success": "scaffold-green",
    "error": "scaffold-red",
    "running": "scaffold-blue",
    "waiting": "scaffold-yellow",
    "killed": "scaffold-orange"
}

var state_icons = {
    "not_started": '<i class="w3-medium fa-regular fa-circle"></i>',
    "success": '<i class="w3-medium fa-solid fa-circle-check"></i>',
    "error": '<i class="w3-medium fa-solid fa-circle-exclamation"></i>',
    "running": '<i class="w3-medium fa-sharp fa-solid fa-spinner fa-spin"></i>',
    "waiting": '<i class="w3-medium fa-solid fa-clock"></i>',
    "killed": '<i class="w3-medium fa-solid fa-skull"></i>'
}

var state_colors_hex = {
    "not_started": "#373F51",
    "success": "#A3BE8C",
    "error": "#BF616A",
    "running": "#5E81AC",
    "waiting": "#EBCB8B",
    "killed": "#D08770"
}

var state_text_colors = {
    "not_started": "scaffold-text-charcoal",
    "success": "scaffold-text-green",
    "error": "scaffold-text-red",
    "running": "scaffold-text-blue",
    "waiting": "scaffold-text-yellow",
    "killed": "scaffold-text-orange"
}

color_keys = ["not_started", "success", "error", "running", "waiting", "killed"]

var hidden = []
var disabled = []

function render() {
    let prefix = $("#search").val();
    prefix = prefix.toLowerCase();

    if (prefix == "") {
        hidden = []
    } else {
        for (let [key, task] of Object.entries(tasks)) {
            if (key.toLowerCase().indexOf(prefix) == -1) {
                hidden.push(key)
            }
        }
    }
    updateNodes()
}

function toggleCurrentState() {
    let sidebar = document.getElementById("current-state")
    if (sidebar.className.indexOf("show") == -1) {
        // Close input
        let input = document.getElementById("current-input")
        input.classList.remove("show");
        $("#current-input").css("left", `calc(100%)`)
        // Close legend
        let legend = document.getElementById("current-legend")
        legend.classList.remove("show");
        $("#current-legend").css("left", `calc(100%)`)
        // Show state
        sidebar.classList.add("show");
        sidebar.classList.remove("right-slide-out-500");
        sidebar.classList.add("right-slide-in-500")
        $("#current-state").css("left", `calc(100% - ${right_panel_width}px)`)
    } else {
        sidebar.classList.remove("show");
        sidebar.classList.remove("right-slide-in-500");
        sidebar.classList.add("right-slide-out-500")
        $("#current-state").css("left", `calc(100%)`)
    }
}

function toggleCurrentInput() {
    let sidebar = document.getElementById("current-input")
    if (sidebar.className.indexOf("show") == -1) {
        // Close state
        let state = document.getElementById("current-state")
        state.classList.remove("show");
        $("#current-state").css("left", `calc(100%)`)
        // Close legend
        let legend = document.getElementById("current-legend")
        legend.classList.remove("show");
        $("#current-legend").css("left", `calc(100%)`)
        // Show input
        sidebar.classList.add("show");
        sidebar.classList.remove("right-slide-out-500");
        sidebar.classList.add("right-slide-in-500")
        $("#current-input").css("left", `calc(100% - ${right_panel_width}px)`)
    } else {
        sidebar.classList.remove("show");
        sidebar.classList.remove("right-slide-in-500");
        sidebar.classList.add("right-slide-out-500")
        $("#current-input").css("left", `calc(100%)`)
    }
}

function toggleCurrentLegend() {
    let sidebar = document.getElementById("current-legend")
    if (sidebar.className.indexOf("show") == -1) {
        // Close state
        let state = document.getElementById("current-state")
        state.classList.remove("show");
        $("#current-state").css("left", `calc(100%)`)
        // Close input
        let input = document.getElementById("current-input")
        input.classList.remove("show");
        $("#current-input").css("left", `calc(100%)`)
        // Show legend
        sidebar.classList.add("show");
        sidebar.classList.remove("right-slide-out-500");
        sidebar.classList.add("right-slide-in-500")
        $("#current-legend").css("left", `calc(100% - ${right_panel_width}px)`)
    } else {
        sidebar.classList.remove("show");
        sidebar.classList.remove("right-slide-in-500");
        sidebar.classList.add("right-slide-out-500")
        $("#current-legend").css("left", `calc(100%)`)
    }
}

function updateDatastore() {
    $("#current-state-div").empty()
    theme = localStorage.getItem('scaffold-theme');

    for (let [key, value] of Object.entries(datastore.env)) {
        shouldShow = true
        for (let idx = 0; idx < inputs.length; idx++) {
            let i = inputs[idx]
            if (i.name == key) {
                if (i.type == "password") {
                    shouldShow = false
                    break
                }
            }
        }
        if (!shouldShow) {
            continue
        }
        html = `<div class="w3-bar-item ${theme} theme-base w3-border-bottom theme-border-light">
            <b>${key}</b>
        </div>
        <div class="w3-bar-item ${theme} theme-light w3-border-bottom theme-border-base">
            ${value}
        </div>`
        $("#current-state-div").append(html)
    }
}

function getDatastore() {
    let parts = window.location.href.split('/')
    let cascadeName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/datastore/" + cascadeName,
        type: "GET",
        contentType: "application/json",
        success: function (result) {
            datastore = result
            updateDatastore()
            
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.href = "/ui/login";
            }
        }
    });
}

function getTasks() {
    let parts = window.location.href.split('/')
    let cascadeName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/task/" + cascadeName,
        type: "GET",
        contentType: "application/json",
        success: function (result) {
            rawTasks = result
            for (let task of rawTasks) {
                let status = getStatusFromName(task.name)
                let extra_icons = ""
                if (task.auto_execute) {
                    extra_icons += '<i class="fa-solid fa-forward w3-medium" style="float:right;margin-right:4px;margin-left:4px;"></i>'
                }
                if (task.cron != "") {
                    extra_icons += '<i class="fa-solid fa-clock w3-medium" style="float:right;margin-right:4px;margin-left:4px;"></i>'
                }

                tasks[task.name] = {
                    "title": {
                        "background_color": state_colors_hex[status],
                        "foreground_color": "#ffffff",
                        "text": `${state_icons[status]}&nbsp;&nbsp;${task.name}${extra_icons}`
                    },
                    "out": {},
                    "func": "changeStateName",
                    "disabled": task.disabled,
                    "extra_icons": extra_icons,
                    "parents": []
                }
            }
            console.log(JSON.stringify(rawTasks))
            console.log(JSON.stringify(tasks))
            for (let task of rawTasks) {
                if (task.depends_on.success != null && task.depends_on.success != undefined && task.depends_on.success.length > 0) {
                    for (let name of task.depends_on.success) {
                        console.log(name)
                        tasks[task.name].parents.push(name)
                        if (tasks[name].out['Success'] !== undefined) {
                            tasks[name].out['Success'].push(task.name)
                        } else {
                            tasks[name].out['Success'] = [task.name]
                        }
                    }
                }
                if (task.depends_on.error != null && task.depends_on.error != undefined && task.depends_on.error.length > 0) {
                    for (let name of task.depends_on.error) {
                        tasks[task.name].parents.push(name)
                        if (tasks[name].out['Error'] !== undefined) {
                            tasks[name].out['Error'].push(task.name)
                        } else {
                            tasks[name].out['Error'] = [task.name]
                        }
                    }
                }
                if (task.depends_on.always != null && task.depends_on.always != undefined && task.depends_on.always.length > 0) {
                    for (let name of task.depends_on.always) {
                        tasks[task.name].parents.push(name)
                        if (tasks[name].out['Always'] !== undefined) {
                            tasks[name].out['Always'].push(task.name)
                        } else {
                            tasks[name].out['Always'] = [task.name]
                        }
                    }
                }
            }
            getStates(true)
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.href = "/ui/login";
            }
        }
    });
}

function getStates(shouldInit) {
    let parts = window.location.href.split('/')
    let cascadeName = parts[parts.length - 1]

    $.ajax({
        url: "/api/v1/state/" + cascadeName,
        type: "GET",
        contentType: "application/json",
        success: function (result) {
            states = result 
            updateNodes()
            if (shouldInit) {
                workflow = new Workflow("cascade-canvas", "cascade-card", "", "", 995, "light theme-light", tasks, pin_colors)
                setInterval(function() {
                    workflow.UpdateWorkflow()
                }, workflowIntervalMilliSeconds);
                setInterval(updateStateStatus, stateIntervalMilliSeconds)
            }
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.href = "/ui/login";
            }
        }
    });
}

function getStatusFromName(n, states) {
    if (states != undefined) {
        for (let state of states) {
            if (state.task == n) {
                return state.status
            }
        }
    }
    return "not_started"
}

function triggerRun() {
    let parts = window.location.href.split('/')
    let cascadeName = parts[parts.length - 1]
    let taskName = CurrentStateName
    
    if (taskName != "") {
        $.ajax({
        url: "/api/v1/run/" + cascadeName + "/" + taskName,
        type: "POST",
        success: function (result) {
            console.log("Run triggered")
        },
        error: function (result) {
            console.log(result)
            if (result.status == 401) {
                window.location.href = "/ui/login";
            }
        }
    });
    }
}

function updateNodes() {
    disabled = []
    for (let state of states) {
        if (state.disabled) {
            disabled.push(state.task)
        }
    }
    for (let [key, task] of Object.entries(tasks)) {
        $(`#${key}`).css("filter", `brightness(100%)`)

        if (hidden.includes(key)) {
            $(`#${key}`).css("filter", `brightness(66%)`)
        }

        if (disabled.includes(key)) {
            $(`#${key}`).css("filter", `brightness(33%)`)
        }
    }
}

function updateStates() {
    if (states != undefined) {
        for (let state of states) {
            if (state.task.startsWith('SCAFFOLD_CHECK') || state.task.startsWith('SCAFFOLD_PREVIOUS')) {
                continue
            }
            let color = state_colors_hex[state.status]
            let icon = state_icons[state.status]
            let extra_icons = tasks[state.task].extra_icons
            let current_color = $(`#${state.task}-header`).css('background-color')
            if (color == current_color) {
                continue
            }
            $(`#${state.task}-header`).css('background-color', color)
            $(`#${state.task}-header-text`).html(`${icon}&nbsp;&nbsp;${state.task}${extra_icons}` )
        }
    }
}

$(document).ready(
    function () {
        $("#current-state").css("left", `calc(100%)`)
        $("#current-input").css("left", `calc(100%)`)
        $("#sidebar").css("left", "-300px")

        getTasks()

        setInterval(function() {
            getStates(false)
        }, stateIntervalMilliSeconds)

        setInterval(function() {
            getDatastore()
        }, stateIntervalMilliSeconds)

        setInterval(function() {
            updateStates()
        }, workflowIntervalMilliSeconds)

        
    }
)