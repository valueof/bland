function main() {
    window.addEventListener("click", (ev) => {
        const target = ev.target
        if (!target) {
            return
        }

        const action = target.getAttribute("data-action")
        if (!action) {
            return
        }

        switch (action) {
            case "mark-read":
                markAsRead(ev)
                break
            default:
                console.error("unsupported action:", action)
        }
    })
}

/** actions */

async function markAsRead(ev) {
    const target = ev.target
    const id = target.getAttribute("data-id")
    
    if (!id) {
        console.error("called markAsRead without id")
        return
    }

    const resp = await fetch('/api/mark-read/', {method: "POST", body: id})
    if (!resp.ok) {
        return
    }

    const parent = target.closest(".bookmarks--markAsRead")
    if (!parent) {
        console.error("couldn't find suitable .bookmarks--markAsRead")
        return
    }

    parent.replaceChild(document.createTextNode("done!"), target)

    setTimeout(() => {
        const actions = parent.closest(".bookmarks--actions")
        if (!actions) {
            console.error("couldn't find suitable .bookmarks--actions")
            return
        }
    
        actions.removeChild(parent)
    }, 2000)
}

window.addEventListener("DOMContentLoaded", main)