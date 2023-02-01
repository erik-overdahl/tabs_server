var port

function reconnect() {
  port = browser.runtime.connectNative("tabs_server")
  port.onDisconnect.addListener(() => {
    if (browser.runtime.lastError) {
      console.log("DISCONNECT", browser.runtime.lastError)
    } else {
      console.log("DISCONNECT: unknown reason")
    }
  })
  port.onMessage.addListener((msg) => { handleMsg(msg) })
  return port
}

port = reconnect()


/*
messages sent from the external server will have
  - `action`(a string)
  - `id` (a number)
  - `content` object to be unpacked in the BrowserTabs method
    specified by `action`
 */

function sendErr(id, err) {
  msg = {
    action: "error",
    id: id,
    content: err.message,
  }
  console.log(msg)
  port.postMessage(msg)
}

function sendSuccess(id, content = {}) {
  port.postMessage({
    action: "success",
    id: id,
    content: content,
  })
}

function handleMsg(msg) {
  console.log("Received:", msg)
  if (!msg.id) {
    sendErr(nil, "Received message with no id")
    return
  }
  if (!msg.action) {
    sendErr(msg.id, "Received message with no action")
    return
  }

  const args = msg.content

  switch (msg.action) {
    case "list":
      // is it possible this could return too much data?
      browser.tabs.query({})
        .then(result => {
          console.log("SENDING: list")
          port.postMessage({
            action: "list",
            id: msg.id,
            content: result
          })})
        .catch(err => sendErr(msg.id, err))

    case "query":
      // is it possible this could return too much data?
      browser.tabs.query(args)
        .then(result =>
          port.postMessage({
            action: "query",
            id: msg.id,
            content: result
          }))
        .catch(err => sendErr(msg.id, err))

    case "create":
      browser.tabs.create(args)
        .then(tab => sendSuccess(msg.id, tab.id))
        .catch(err => sendErr(msg.id, err))

    case "duplicate":
      browser.tabs.duplicate(args.tabId, args.props)
        .then(tab => sendSuccess(msg.id, tab.id))
        .catch(err => sendErr(msg.id, err))

    case "update":
      browser.tabs.update(args.tabId, args.delta)
        .then(sendSuccess(msg.id))
        .catch(err => sendErr(msg.id, err))

    case "move":
      browser.tabs.move(args.tabId, args.props)
        .then(sendSuccess(msg.id))
        .catch(err => sendErr(msg.id, err))

    case "reload":
      browser.tabs.reload(args.tabId, { bypassCache: args.bypassCache })
        .then(sendSuccess(msg.id))
        .catch(err => sendErr(msg.id, err))

    case "remove":
      browser.tabs.remove(args)
        .then(sendSuccess(msg.id))
        .catch(err => sendErr(msg.id, err))

    case "discard":
      browser.tabs.discard(args)
        .then(sendSuccess(msg.id))
        .catch(err => sendErr(msg.id, err))

    // requires "tabHide" permission
    case "hide":
      browser.tabs.hide(args)
        .then(sendSuccess(msg.id))
        .catch(err => sendErr(msg.id, err))

    case "show":
      browser.tabs.show(args)
        .then(sendSuccess(msg.id))
        .catch(err => sendErr(msg.id, err))

    case "toggleReaderMode":
      browser.tabs.toggleReaderMode(args)
        .then(sendSuccess(msg.id))
        .catch(err => sendErr(msg.id, err))

    case "goForward":
      browser.tabs.goForward(args)
        .then(sendSuccess(msg.id))
        .catch(err => sendErr(msg.id, err))

    case "goBack":
      browser.tabs.goBack(args)
        .then(sendSuccess(msg.id))
        .catch(err => sendErr(msg.id, err))

    default:
      sendErr(msg.id, `Action ${msg.Action} is unknown`)
  }
// captureTab
// captureVisibleTab
// connect
// detectLanguage
// executeScript
// get
// getAllInWindow
// getCurrent
// getSelected
// getZoom
// getZoomSettings
}

/* activeInfo {tabId, previousTabId, windowId} */
browser.tabs.onActivated.addListener(
  (activeInfo) => {
    console.log("Tab activated:", activeInfo)
    port.postMessage({
      action: "activated",
      content: {
        tabId: activeInfo.tabId,
        previous: activeInfo.previousTabId,
        windowId: activeInfo.windowId
      }
    })
  }
)

/* changeInfo contains the tab properties that changed */
browser.tabs.onUpdated.addListener(
  (tabId, changeInfo, tab) => {
    console.log("Tab updated:", tabId, changeInfo)
    port.postMessage({
      action: "updated",
      content: {
        tabId: tabId,
        delta: changeInfo,
      }
    })
  }
)

browser.tabs.onCreated.addListener(
  (tab) => {
    console.log("Tab created:", tab)
    port.postMessage({
      action: "created",
      content: tab,
    })
  }
)

/* moveInfo {windowId, fromIndex, toIndex} */
browser.tabs.onMoved.addListener(
  (tabId, moveInfo) => {
    console.log("Tab moved:", tabId, moveInfo)
    port.postMessage({
      action: "moved",
      content: {
        tabId: tabId,
        windowId: moveInfo.windowId,
        fromIndex: moveInfo.fromIndex,
        toIndex: moveInfo.toIndex,
      }
    })
  }
)

/* removeInfo {windowId, isWindowClosing} */
browser.tabs.onRemoved.addListener(
  (tabId, removeInfo) => {
    console.log("Tab removed:", tabId, removeInfo)
    port.postMessage({
      action: "removed",
      content: {
        tabId: tabId,
        windowId: removeInfo.windowId,
        isWindowClosing: removeInfo.isWindowClosing,
      }
    })
  }
)

/* attachInfo {newWindowId, newPosition} */
browser.tabs.onAttached.addListener(
  (tabId, info) => {
    console.log("Tab attached:", tabId, info)
    port.postMessage({
      action: "attached",
      content: {
        tabId: tabId,
        windowId: info.oldWindowId,
        position: info.oldPosition,
      }
    })
  }
)

/* detachInfo {oldWindowId, oldPosition} */
browser.tabs.onDetached.addListener(
  (tabId, info) => {
    console.log("Tab detached:", tabId, info)
    port.postMessage({
      action: "detached",
      content: {
        tabId: tabId,
        windowId: info.oldWindowId,
        position: info.oldPosition,
      }
    })
  }
)


/*

// highlightInfo {windowId, tabIds}
browser.tabs.onHighlighted.addListener(
  (highlightInfo) => {

  }
)

// zoomChangeInfo {tabId, oldZoomFactor, newZoomFactor, zoomSettings}
browser.tabs.onZoomChange.addListener(
  (zoomChangeInfo) => {

  }
)

*/
