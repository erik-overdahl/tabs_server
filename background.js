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

function handleMsg(msg) {
  if (msg.type === "request") {
      handleRequest(msg.data)
  } else {
      console.log("ERROR: unexpected msg from gateway:", msg)
      sendErr(nil, "ERROR: unexpected msg from gateway:" + JSON.stringify(msg))
  }
}

/*
messages sent from the external server will have
  - `action`(a string)
  - `id` (a number)
  - `content` object to be unpacked in the BrowserTabs method
    specified by `action`
 */

function sendResponse(id, status, info) {
  msg = {
    type: "response",
    data: {
      id: id,
      status: status,
      info: info,
    }}
  console.log("SENDING:", msg)
  return port.postMessage(msg)
}

function sendErr(id, message) {
  sendResponse(id, "error", message)
}

function sendSuccess(id, content = {}) {
  sendResponse(id, "success", content)
}

// request: {id: uuid, tabId: Optional[int], tabIds: Optional[[]int], props: Optional[{}]}
function handleRequest(request) {
  console.log("Received:", request)
  if (!request.id) {
    sendErr(nil, "Received message with no id")
    return
  }
  if (!request.method) {
    sendErr(request.id, "Received message with no action")
    return
  }

  switch (request.method) {
    case "list":
      // is it possible this could return too much data?
      browser.tabs.query({})
        .then((result) => { sendResponse(request.id, "list", result) })
        .catch(err => sendErr(request.id, err.message))
      break

    case "query":
      // is it possible this could return too much data?
      browser.tabs.query(request.props)
        .then((result) => { sendResponse(request.id, "query", result) })
        .catch(err => sendErr(request.id, err.message))
      break

    case "create":
      browser.tabs.create(request.props)
        .then((tab) => { sendSuccess(request.id, tab.id) })
        .catch(err => sendErr(request.id, err.message))
      break

    case "duplicate":
      browser.tabs.duplicate(request.tabId, request.props)
        .then((tab) => { sendSuccess(request.id, tab.id) })
        .catch(err => sendErr(request.id, err.message))
      break

    case "update":
      browser.tabs.update(request.tabId, request.props)
        .then(() => { sendSuccess(request.id) })
        .catch(err => sendErr(request.id, err.message))
      break

    case "move":
      browser.tabs.move(request.tabId, request.props)
        .then(() => { sendSuccess(request.id) })
        .catch(err => sendErr(request.id, err.message))
      break

    case "reload":
      browser.tabs.reload(request.tabId, request.props)
        .then(() => { sendSuccess(request.id) })
        .catch(err => sendErr(request.id, err.message))
      break

    case "remove":
      browser.tabs.remove(request.tabIds)
        .then(() => { sendSuccess(request.id) })
        .catch((err) => { sendErr(request.id, err.message) })
      break

    case "discard":
      browser.tabs.discard(request.tabIds)
        .then(() => { sendSuccess(request.id) })
        .catch((err) => { sendErr(request.id, err.message) })
      break

    // requires "tabHide" permission
    case "hide":
      browser.tabs.hide(request.tabIds)
        .then(() => { sendSuccess(request.id) })
        .catch(err => sendErr(request.id, err.message))
      break

    case "show":
      browser.tabs.show(request.tabIds)
        .then(() => sendSuccess(request.id))
        .catch(err => sendErr(request.id, err.message))
      break

    case "toggleReaderMode":
      browser.tabs.toggleReaderMode(request.tabId)
        .then(() => sendSuccess(request.id))
        .catch(err => sendErr(request.id, err.message))
      break

    case "goForward":
      browser.tabs.goForward(request.tabId)
        .then(() => sendSuccess(request.id))
        .catch(err => sendErr(request.id, err.message))
      break

    case "goBack":
      browser.tabs.goBack(request.tabId)
        .then(() => sendSuccess(request.id))
        .catch(err => sendErr(request.id, err.message))
      break

    default:
      sendErr(request.id, `Action ${request.Action} is unknown`)
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

function sendEvent(type, data) {
  msg = {
    type: "event",
    data: {
      type: type,
      data: data,
    }
  }
  console.log(`EVENT: ${type}:`, data)
  port.postMessage(msg)
}

/* activeInfo {tabId, previousTabId, windowId} */
browser.tabs.onActivated.addListener(
  (activeInfo) => sendEvent(
    "activated",
    {
      tabId: activeInfo.tabId,
      previous: activeInfo.previousTabId,
      windowId: activeInfo.windowId
    })
)

/* changeInfo contains the tab properties that changed */
browser.tabs.onUpdated.addListener(
  (tabId, changeInfo, tab) => sendEvent(
    "updated",
    {
      tabId: tabId,
      delta: changeInfo,
    })
)

browser.tabs.onCreated.addListener(
  (tab) => sendEvent("created", tab))

/* moveInfo {windowId, fromIndex, toIndex} */
browser.tabs.onMoved.addListener(
  (tabId, moveInfo) => sendEvent(
    "moved",
    {
      tabId: tabId,
      windowId: moveInfo.windowId,
      fromIndex: moveInfo.fromIndex,
      toIndex: moveInfo.toIndex,
    })
)

/* removeInfo {windowId, isWindowClosing} */
browser.tabs.onRemoved.addListener(
  (tabId, removeInfo) => sendEvent(
    "removed",
    {
      tabId: tabId,
      windowId: removeInfo.windowId,
      isWindowClosing: removeInfo.isWindowClosing,
    })
)

/* attachInfo {newWindowId, newPosition} */
browser.tabs.onAttached.addListener(
  (tabId, info) => sendEvent(
    "attached",
    {
      tabId: tabId,
      windowId: info.oldWindowId,
      position: info.oldPosition,
    })
)

/* detachInfo {oldWindowId, oldPosition} */
browser.tabs.onDetached.addListener(
  (tabId, info) => sendEvent(
    "detached",
    {
      tabId: tabId,
      windowId: info.oldWindowId,
      position: info.oldPosition,
    })
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
