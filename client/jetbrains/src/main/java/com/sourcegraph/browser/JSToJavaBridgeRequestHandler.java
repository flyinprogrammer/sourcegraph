package com.sourcegraph.browser;

import com.google.gson.JsonObject;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.intellij.ui.jcef.JBCefJSQuery;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.ThemeUtil;
import com.sourcegraph.find.FindPopupPanel;
import com.sourcegraph.find.FindService;
import com.sourcegraph.find.PreviewContent;
import com.sourcegraph.find.Search;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.io.PrintWriter;
import java.io.StringWriter;
import java.time.Instant;
import java.time.format.DateTimeFormatter;
import java.util.Date;

public class JSToJavaBridgeRequestHandler {
    private final Project project;
    private final FindPopupPanel findPopupPanel;
    private final FindService findService;

    public JSToJavaBridgeRequestHandler(@NotNull Project project, @NotNull FindPopupPanel findPopupPanel, @NotNull FindService findService) {
        this.project = project;
        this.findPopupPanel = findPopupPanel;
        this.findService = findService;
    }

    public JBCefJSQuery.Response handle(@NotNull JsonObject request) {
        String action = request.get("action").getAsString();
        JsonObject arguments;
        PreviewContent previewContent;
        try {
            switch (action) {
                case "getConfig":
                    return createSuccessResponse(ConfigUtil.getConfigAsJson(project));
                case "getTheme":
                    JsonObject currentThemeAsJson = ThemeUtil.getCurrentThemeAsJson();
                    return createSuccessResponse(currentThemeAsJson);
                case "saveLastSearch":
                    arguments = request.getAsJsonObject("arguments");
                    String query = arguments.get("query").getAsString();
                    boolean caseSensitive = arguments.get("caseSensitive").getAsBoolean();
                    String patternType = arguments.get("patternType").getAsString();
                    String selectedSearchContextSpec = arguments.get("selectedSearchContextSpec").getAsString();
                    ConfigUtil.setLastSearch(project, new Search(
                        query,
                        caseSensitive,
                        patternType,
                        selectedSearchContextSpec
                    ));
                    return createSuccessResponse(new JsonObject());
                case "loadLastSearch":
                    Search lastSearch = ConfigUtil.getLastSearch(this.project);

                    if (lastSearch == null) {
                        return createSuccessResponse(null);
                    }

                    JsonObject lastSearchAsJson = new JsonObject();
                    lastSearchAsJson.addProperty("query", lastSearch.getQuery());
                    lastSearchAsJson.addProperty("caseSensitive", lastSearch.isCaseSensitive());
                    lastSearchAsJson.addProperty("patternType", lastSearch.getPatternType());
                    lastSearchAsJson.addProperty("selectedSearchContextSpec", lastSearch.getSelectedSearchContextSpec());
                    return createSuccessResponse(lastSearchAsJson);
                case "previewLoading":
                    arguments = request.getAsJsonObject("arguments");
                    // Wait a bit to avoid flickering in case of a fast network
                    new Thread(() -> {
                        try {
                            Thread.sleep(300);
                        } catch (InterruptedException ignored) {
                        }
                        ApplicationManager.getApplication().invokeLater(() -> findPopupPanel.indicateLoadingIfInTime(Date.from(
                            Instant.from(DateTimeFormatter.ISO_INSTANT.parse(arguments.get("timeAsISOString").getAsString())))));
                    }).start();
                    return createSuccessResponse(null);
                case "preview":
                    arguments = request.getAsJsonObject("arguments");
                    previewContent = PreviewContent.fromJson(project, arguments);
                    ApplicationManager.getApplication().invokeLater(() -> findPopupPanel.setPreviewContentIfInTime(previewContent));
                    return createSuccessResponse(null);
                case "clearPreview":
                    arguments = request.getAsJsonObject("arguments");
                    ApplicationManager.getApplication().invokeLater(() -> findPopupPanel.clearPreviewContentIfInTime(Date.from(
                        Instant.from(DateTimeFormatter.ISO_INSTANT.parse(arguments.get("timeAsISOString").getAsString())))));
                    return createSuccessResponse(null);
                case "open":
                    arguments = request.getAsJsonObject("arguments");
                    try {
                        previewContent = PreviewContent.fromJson(project, arguments);
                    } catch (Exception e) {
                        return createErrorResponse("Parsing error while opening link: " + e.getClass().getName() + ": " + e.getMessage(), convertStackTraceToString(e));
                    }

                    ApplicationManager.getApplication().invokeLater(() -> {
                        try {
                            previewContent.openInEditorOrBrowser();
                        } catch (Exception e) {
                            Logger logger = Logger.getInstance(JSToJavaBridgeRequestHandler.class);
                            logger.warn("Error while opening link.", e);
                        }
                    });
                    return createSuccessResponse(null);
                case "indicateFinishedLoading":
                    arguments = request.getAsJsonObject("arguments");
                    ApplicationManager.getApplication().invokeLater(() -> findPopupPanel.indicateAuthenticationStatus(arguments.get("wasServerAccessSuccessful").getAsBoolean(), arguments.get("wasAuthenticationSuccessful").getAsBoolean()));
                    return createSuccessResponse(null);
                case "windowClose":
                    ApplicationManager.getApplication().invokeLater(findService::hidePopup);
                    return createSuccessResponse(null);
                default:
                    return createErrorResponse("Unknown action: '" + action + "'.", "No stack trace");
            }
        } catch (Exception e) {
            return createErrorResponse(action + ": " + e.getClass().getName() + ": " + e.getMessage(), convertStackTraceToString(e));
        }
    }

    @NotNull
    public Project getProject() {
        return project;
    }

    public JBCefJSQuery.Response handleInvalidRequest(@NotNull Exception e) {
        return createErrorResponse("Invalid JSON passed to bridge. The error is: " + e.getClass() + ": " + e.getMessage(), convertStackTraceToString(e));
    }

    @NotNull
    private JBCefJSQuery.Response createSuccessResponse(@Nullable JsonObject result) {
        return new JBCefJSQuery.Response(result != null ? result.toString() : "null");
    }

    @NotNull
    private JBCefJSQuery.Response createErrorResponse(@NotNull String errorMessage, @NotNull String stackTrace) {
        return new JBCefJSQuery.Response(null, 0, errorMessage + "\n" + stackTrace);
    }

    @NotNull
    private String convertStackTraceToString(@NotNull Exception e) {
        StringWriter sw = new StringWriter();
        PrintWriter pw = new PrintWriter(sw);
        e.printStackTrace(pw);
        return sw.toString();
    }
}
