const state = {
  sourceMode: "testcase",
  testcases: [],
  selectedTestcase: "",
  testcaseContent: "",
  previewPuzzle: null,
  solveResponse: null,
  currentStep: 0,
  autoplayId: null,
  toastId: null,
};

const elements = {
  toast: document.getElementById("toast"),
  sourceButtons: Array.from(document.querySelectorAll("[data-source-mode]")),
  testcasePanel: document.getElementById("panel-testcase"),
  manualPanel: document.getElementById("panel-manual"),
  testcaseSelect: document.getElementById("testcase-select"),
  loadTestcase: document.getElementById("load-testcase"),
  testcasePreview: document.getElementById("testcase-preview"),
  manualInput: document.getElementById("manual-input"),
  algorithmSelect: document.getElementById("algorithm-select"),
  heuristicSelect: document.getElementById("heuristic-select"),
  runMain: document.getElementById("run-main"),
  refreshLayout: document.getElementById("refresh-layout"),
  fileInput: document.getElementById("file-input"),
  pickFile: document.getElementById("pick-file"),
  boardGrid: document.getElementById("board-grid"),
  boardCaption: document.getElementById("board-caption"),
  visualizerSubtitle: document.getElementById("visualizer-subtitle"),
  resetFrames: document.getElementById("reset-frames"),
  playToggle: document.getElementById("play-toggle"),
  playIcon: document.getElementById("play-icon"),
  stepPrev: document.getElementById("step-prev"),
  stepNext: document.getElementById("step-next"),
  stepLabel: document.getElementById("step-label"),
  gotoStep: document.getElementById("goto-step"),
  timeline: document.getElementById("timeline"),
  frameDirection: document.getElementById("frame-direction"),
  frameCost: document.getElementById("frame-cost"),
  frameActor: document.getElementById("frame-actor"),
  frameCheckpoints: document.getElementById("frame-checkpoints"),
  resultStatus: document.getElementById("result-status"),
  resultsSubtitle: document.getElementById("results-subtitle"),
  solutionPath: document.getElementById("solution-path"),
  metricCost: document.getElementById("metric-cost"),
  metricTime: document.getElementById("metric-time"),
  metricIter: document.getElementById("metric-iter"),
  metricMoves: document.getElementById("metric-moves"),
  metricAlgorithm: document.getElementById("metric-algorithm"),
  metricHeuristic: document.getElementById("metric-heuristic"),
  saveSolution: document.getElementById("save-solution"),
};

async function init() {
  bindEvents();
  updateSourceMode("testcase");
  updateHeuristicAvailability();
  updatePlaybackControls();
  renderResultState();

  try {
    await loadTestcases();
  } catch (error) {
    showToast(error.message || "Failed to load testcase list.");
  }
}

function bindEvents() {
  elements.sourceButtons.forEach((button) => {
    button.addEventListener("click", () => updateSourceMode(button.dataset.sourceMode));
  });

  elements.loadTestcase.addEventListener("click", loadSelectedTestcase);
  elements.testcaseSelect.addEventListener("change", () => {
    state.selectedTestcase = elements.testcaseSelect.value;
  });

  elements.algorithmSelect.addEventListener("change", updateHeuristicAvailability);
  elements.runMain.addEventListener("click", runSolver);
  elements.refreshLayout.addEventListener("click", async () => {
    try {
      await loadTestcases();
      showToast("Testcase list refreshed.");
    } catch (error) {
      showToast(error.message || "Unable to refresh testcase list.");
    }
  });

  elements.pickFile.addEventListener("click", () => elements.fileInput.click());
  elements.fileInput.addEventListener("change", handleFileSelection);

  elements.resetFrames.addEventListener("click", () => setCurrentStep(0));
  elements.stepPrev.addEventListener("click", () => setCurrentStep(Math.max(0, state.currentStep - 1)));
  elements.stepNext.addEventListener("click", () => {
    const max = getFrameCount() - 1;
    setCurrentStep(Math.min(max, state.currentStep + 1));
  });
  elements.playToggle.addEventListener("click", togglePlayback);
  elements.gotoStep.addEventListener("change", () => setCurrentStep(Number(elements.gotoStep.value) || 0));
  elements.timeline.addEventListener("input", () => setCurrentStep(Number(elements.timeline.value)));

  elements.saveSolution.addEventListener("click", saveSolutionAsText);
}

async function apiFetch(url, options = {}) {
  const response = await fetch(url, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    ...options,
  });

  const contentType = response.headers.get("content-type") || "";
  const payload = contentType.includes("application/json")
    ? await response.json()
    : await response.text();

  if (!response.ok) {
    const message = typeof payload === "string" ? payload : payload.error || "Request failed";
    throw new Error(message);
  }

  return payload;
}

async function loadTestcases() {
  const testcases = await apiFetch("/api/testcases", { method: "GET" });
  state.testcases = Array.isArray(testcases) ? testcases : [];

  elements.testcaseSelect.innerHTML = "";

  if (state.testcases.length === 0) {
    const option = new Option("No testcase found", "");
    elements.testcaseSelect.add(option);
    elements.testcaseSelect.disabled = true;
    elements.loadTestcase.disabled = true;
    return;
  }

  state.testcases.forEach((item) => {
    const option = new Option(item.name, item.name);
    elements.testcaseSelect.add(option);
  });

  elements.testcaseSelect.disabled = false;
  elements.loadTestcase.disabled = false;
  state.selectedTestcase = state.selectedTestcase || state.testcases[0].name;
  elements.testcaseSelect.value = state.selectedTestcase;
  await loadSelectedTestcase();
}

async function loadSelectedTestcase() {
  const name = elements.testcaseSelect.value;
  if (!name) {
    showToast("Select a testcase first.");
    return;
  }

  const response = await apiFetch(`/api/testcase?name=${encodeURIComponent(name)}`, { method: "GET" });
  state.selectedTestcase = name;
  state.testcaseContent = response.content || "";
  state.previewPuzzle = response.puzzle || null;
  elements.testcasePreview.value = state.testcaseContent;
  renderPreviewPuzzle();
  showToast(`Loaded testcase ${name}.`);
}

function updateSourceMode(mode) {
  state.sourceMode = mode;
  elements.sourceButtons.forEach((button) => {
    button.classList.toggle("is-active", button.dataset.sourceMode === mode);
  });
  elements.testcasePanel.classList.toggle("hidden", mode !== "testcase");
  elements.manualPanel.classList.toggle("hidden", mode !== "manual");

  if (mode === "testcase" && state.previewPuzzle) {
    renderPreviewPuzzle();
  } else if (mode === "manual" && state.solveResponse?.puzzle) {
    renderBoard(state.solveResponse.puzzle, getCurrentFrame());
  }
}

function updateHeuristicAvailability() {
  const algorithm = elements.algorithmSelect.value;
  const requiresHeuristic = ["GBFS", "A*", "IDA*"].includes(algorithm);
  elements.heuristicSelect.disabled = !requiresHeuristic;
  elements.heuristicSelect.closest(".field").style.opacity = requiresHeuristic ? "1" : "0.55";
}

async function handleFileSelection(event) {
  const [file] = Array.from(event.target.files || []);
  if (!file) {
    return;
  }
  await applyPuzzleFile(file);
  event.target.value = "";
}

async function applyPuzzleFile(file) {
  if (!file.name.toLowerCase().endsWith(".txt")) {
    showToast("Only .txt puzzle files are supported.");
    return;
  }

  const text = await file.text();
  updateSourceMode("manual");
  elements.manualInput.value = text;
  showToast(`Loaded ${file.name} into manual input.`);
}

async function runSolver() {
  stopPlayback();

  const payload = {
    testcase: "",
    puzzleText: "",
    algorithm: elements.algorithmSelect.value,
    heuristic: elements.heuristicSelect.disabled ? "" : elements.heuristicSelect.value,
  };

  if (state.sourceMode === "testcase") {
    payload.testcase = elements.testcaseSelect.value;
  } else {
    payload.puzzleText = elements.manualInput.value.trim();
  }

  if (!payload.testcase && !payload.puzzleText) {
    showToast("Puzzle input is still empty.");
    return;
  }

  setBusyState(true);

  try {
    const response = await apiFetch("/api/solve", {
      method: "POST",
      body: JSON.stringify(payload),
    });

    state.solveResponse = response;
    state.previewPuzzle = response.puzzle;
    state.currentStep = 0;
    renderPreviewPuzzle();
    renderResultState();
    updatePlaybackControls();
    showToast(response.result.found ? "Solver finished successfully." : "Solver finished without a solution.");
  } catch (error) {
    state.solveResponse = null;
    renderResultState(error.message || "Solver request failed.");
    showToast(error.message || "Solver request failed.");
  } finally {
    setBusyState(false);
  }
}

function renderPreviewPuzzle() {
  if (!state.previewPuzzle) {
    renderEmptyBoard("Waiting for puzzle data.");
    updateMetaSnapshot(null);
    return;
  }

  const frame = state.solveResponse ? getCurrentFrame() : null;
  renderBoard(state.previewPuzzle, frame);
  updateMetaSnapshot(state.previewPuzzle);

  if (!state.solveResponse) {
    elements.visualizerSubtitle.textContent = "Previewing the currently loaded puzzle.";
    elements.boardCaption.textContent = "Board preview before solver execution.";
  }
}

function renderBoard(puzzle, frame) {
  elements.boardGrid.innerHTML = "";
  elements.boardGrid.style.gridTemplateColumns = `repeat(${puzzle.cols}, minmax(0, 1fr))`;

  const actor = frame?.state?.actor || puzzle.start;
  const pathLookup = new Set((frame?.path || []).map((pos) => `${pos.row},${pos.col}`));

  for (let row = 0; row < puzzle.rows; row += 1) {
    for (let col = 0; col < puzzle.cols; col += 1) {
      const token = puzzle.board[row][col];
      const cost = puzzle.costs[row][col];
      const cell = document.createElement("div");
      cell.className = `board-cell ${getTileClass(token)}`;

      if (pathLookup.has(`${row},${col}`)) {
        cell.classList.add("board-cell-visited");
      }
      if (actor.row === row && actor.col === col) {
        cell.classList.add("board-cell-actor");
      }

      const tokenEl = document.createElement("span");
      tokenEl.className = "board-cell-token";
      tokenEl.textContent = getTokenLabel(token, row, col, puzzle);
      cell.appendChild(tokenEl);

      const costEl = document.createElement("span");
      costEl.className = "board-cell-cost";
      costEl.textContent = String(cost);
      cell.appendChild(costEl);

      cell.title = `(${row},${col}) token=${token} cost=${cost}`;
      elements.boardGrid.appendChild(cell);
    }
  }

  const frameTitle = frame?.title || "Board Preview";
  elements.visualizerSubtitle.textContent = frame ? frameTitle : "Previewing the currently loaded puzzle.";
  elements.boardCaption.textContent = frame
    ? `${frameTitle} | Actor at ${formatPosition(actor)}`
    : `Puzzle ${puzzle.rows}x${puzzle.cols} ready for execution.`;
  renderFrameInfo(frame, puzzle);
}

function renderFrameInfo(frame, puzzle) {
  if (!frame) {
    elements.frameDirection.textContent = "-";
    elements.frameCost.textContent = "-";
    elements.frameActor.textContent = puzzle ? formatPosition(puzzle.start) : "-";
    elements.frameCheckpoints.textContent = puzzle ? `0 / ${puzzle.checkpointOrder.length}` : "-";
    return;
  }

  const totalCheckpoints = puzzle.checkpointOrder.length;
  elements.frameDirection.textContent = frame.direction || "Start";
  elements.frameCost.textContent = String(frame.pathCost ?? 0);
  elements.frameActor.textContent = formatPosition(frame.state.actor);
  elements.frameCheckpoints.textContent = `${frame.state.nextCheckpoint} / ${totalCheckpoints}`;
}

function renderEmptyBoard(message) {
  elements.boardGrid.innerHTML = "";
  elements.visualizerSubtitle.textContent = "No puzzle loaded yet.";
  elements.boardCaption.textContent = message;
  renderFrameInfo(null, null);
}

function updateMetaSnapshot(puzzle) {
  void puzzle;
}

function renderResultState(errorMessage = "") {
  const response = state.solveResponse;
  const result = response?.result;

  if (!response || !result) {
    elements.resultStatus.textContent = errorMessage ? "Error" : "Belum jalan";
    elements.resultStatus.className = `status-pill${errorMessage ? " error" : ""}`;
    elements.resultsSubtitle.textContent = errorMessage || "Belum dijalankan.";
    elements.solutionPath.textContent = "-";
    elements.metricCost.textContent = "-";
    elements.metricTime.textContent = "-";
    elements.metricIter.textContent = "-";
    elements.metricMoves.textContent = "-";
    elements.metricAlgorithm.textContent = "-";
    elements.metricHeuristic.textContent = "-";
    elements.saveSolution.disabled = true;
    return;
  }

  elements.resultStatus.textContent = result.found ? "Solusi ditemukan!" : "Tidak ada solusi";
  elements.resultStatus.className = `status-pill${result.found ? " success" : " error"}`;
  elements.resultsSubtitle.textContent = result.found
    ? "Rute ditemukan dari input saat ini."
    : "Solver selesai, tapi rute valid tidak ditemukan.";
  elements.solutionPath.textContent = result.moves || "(no moves)";
  elements.metricCost.textContent = String(result.totalCost);
  elements.metricTime.textContent = `${result.durationMs} ms`;
  elements.metricIter.textContent = String(result.iterations);
  elements.metricMoves.textContent = String(result.moveCount);
  elements.metricAlgorithm.textContent = result.algorithm || "-";
  elements.metricHeuristic.textContent = result.heuristic || "Not required";
  elements.saveSolution.disabled = false;
}

function updatePlaybackControls() {
  const frameCount = getFrameCount();
  const max = Math.max(0, frameCount - 1);
  elements.timeline.max = String(max);
  elements.timeline.value = String(Math.min(state.currentStep, max));
  elements.gotoStep.max = String(max);
  elements.gotoStep.value = String(Math.min(state.currentStep, max));
  elements.stepLabel.textContent = `STEP ${Math.min(state.currentStep, max)} / ${max}`;
  elements.stepPrev.disabled = frameCount <= 1 || state.currentStep <= 0;
  elements.stepNext.disabled = frameCount <= 1 || state.currentStep >= max;
  elements.playToggle.disabled = frameCount <= 1;

  const progress = max === 0 ? 0 : (state.currentStep / max) * 100;
  elements.timeline.style.setProperty("--progress", `${progress}%`);
}

function setCurrentStep(stepIndex) {
  const max = Math.max(0, getFrameCount() - 1);
  state.currentStep = Math.min(Math.max(stepIndex, 0), max);
  updatePlaybackControls();

  if (state.previewPuzzle) {
    renderBoard(state.previewPuzzle, getCurrentFrame());
  }
}

function togglePlayback() {
  if (state.autoplayId) {
    stopPlayback();
    return;
  }

  if (getFrameCount() <= 1) {
    return;
  }

  elements.playIcon.textContent = "pause";
  state.autoplayId = window.setInterval(() => {
    const max = getFrameCount() - 1;
    if (state.currentStep >= max) {
      stopPlayback();
      return;
    }
    setCurrentStep(state.currentStep + 1);
  }, 800);
}

function stopPlayback() {
  if (state.autoplayId) {
    window.clearInterval(state.autoplayId);
    state.autoplayId = null;
  }
  elements.playIcon.textContent = "play_arrow";
}

function setBusyState(isBusy) {
  [elements.runMain].forEach((button) => {
    button.disabled = isBusy;
  });
  if (isBusy) {
    elements.runMain.querySelector("span:last-child").textContent = "Solving...";
  } else {
    elements.runMain.querySelector("span:last-child").textContent = "Solve";
  }
}

function getFrameCount() {
  return state.solveResponse?.frames?.length || 0;
}

function getCurrentFrame() {
  return state.solveResponse?.frames?.[state.currentStep] || null;
}

function getTileClass(token) {
  if (token === "X") {
    return "board-cell-wall";
  }
  if (token === "Z") {
    return "board-cell-start";
  }
  if (token === "O") {
    return "board-cell-goal";
  }
  if (token === "L") {
    return "board-cell-lava";
  }
  if (/^\d+$/.test(token)) {
    return "board-cell-checkpoint";
  }
  return "board-cell-floor";
}

function getTokenLabel(token, row, col, puzzle) {
  if (token === "." || token === "*" || token === "-" || token === "_") {
    return "";
  }
  if (token === "Z") {
    return "S";
  }
  if (token === "O") {
    return "G";
  }
  if (token === "X") {
    return "";
  }
  if (token === "L") {
    return "L";
  }
  if (/^\d+$/.test(token)) {
    return token;
  }

  if (puzzle.start.row === row && puzzle.start.col === col) {
    return "S";
  }
  return token;
}

function formatPosition(position) {
  if (!position) {
    return "-";
  }
  return `(${position.row}, ${position.col})`;
}

function saveSolutionAsText() {
  if (!state.solveResponse?.textOutput) {
    showToast("No solution data to save yet.");
    return;
  }

  downloadFile("solution.txt", state.solveResponse.textOutput, "text/plain;charset=utf-8");
}

function downloadFile(filename, content, mimeType) {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
}

function showToast(message) {
  if (!message) {
    return;
  }

  elements.toast.textContent = message;
  elements.toast.classList.add("is-visible");
  window.clearTimeout(state.toastId);
  state.toastId = window.setTimeout(() => {
    elements.toast.classList.remove("is-visible");
  }, 2800);
}

init();
