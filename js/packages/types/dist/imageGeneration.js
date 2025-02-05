export var GenerationStatus;
(function (GenerationStatus) {
    GenerationStatus[GenerationStatus["Unidentified"] = 0] = "Unidentified";
    GenerationStatus[GenerationStatus["Pending"] = 1] = "Pending";
    GenerationStatus[GenerationStatus["InProgress"] = 2] = "InProgress";
    GenerationStatus[GenerationStatus["GenerateImageOptions"] = 3] = "GenerateImageOptions";
    GenerationStatus[GenerationStatus["Complete"] = 4] = "Complete";
    GenerationStatus[GenerationStatus["Failed"] = 5] = "Failed";
})(GenerationStatus || (GenerationStatus = {}));
export var GenerationBackend;
(function (GenerationBackend) {
    GenerationBackend[GenerationBackend["Unidentified"] = 0] = "Unidentified";
    GenerationBackend[GenerationBackend["Imagine"] = 1] = "Imagine";
    GenerationBackend[GenerationBackend["UseApi"] = 2] = "UseApi";
})(GenerationBackend || (GenerationBackend = {}));
