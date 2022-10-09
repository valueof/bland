window.addEventListener("DOMContentLoaded", main)

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
            case "delete-bookmark":
                deleteBookmark(ev)
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

    const resp = await fetch('/api/mark-read', {method: "POST", body: id})
    if (!resp.ok) {
        return
    }

    const parent = target.closest(".bookmarks--markAsRead")
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

async function deleteBookmark(ev) {
    const target = ev.target
    const id = target.getAttribute("data-id")

    if (!id) {
        console.error("called markAsRead without id")
        return
    }

    const resp = await fetch('/api/delete-bookmark', {method: "POST", body: id})
    if (!resp.ok) {
        return
    }

    const bookmark = target.closest(`#bookmark-${id}`)
    const deleted = document.createElement("div")
    deleted.className = "bookmarks--bookmark u-pill"
    deleted.innerHTML = "bookmark deleted!"
    bookmark.parentNode.replaceChild(deleted, bookmark)

    setTimeout(() => {
        deleted.parentNode.removeChild(deleted)
    }, 2000)
}