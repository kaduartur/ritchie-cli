package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ZupIT/ritchie-cli/pkg/api"
	"github.com/ZupIT/ritchie-cli/pkg/autocomplete"
	"github.com/ZupIT/ritchie-cli/pkg/cmd"
	"github.com/ZupIT/ritchie-cli/pkg/credential/credteam"
	"github.com/ZupIT/ritchie-cli/pkg/env"
	"github.com/ZupIT/ritchie-cli/pkg/env/envcredential"
	"github.com/ZupIT/ritchie-cli/pkg/formula"
	"github.com/ZupIT/ritchie-cli/pkg/metrics"
	"github.com/ZupIT/ritchie-cli/pkg/rcontext"
	"github.com/ZupIT/ritchie-cli/pkg/security"
	"github.com/ZupIT/ritchie-cli/pkg/security/secteam"
	"github.com/ZupIT/ritchie-cli/pkg/session"
	"github.com/ZupIT/ritchie-cli/pkg/session/sessteam"
	"github.com/ZupIT/ritchie-cli/pkg/workspace"
	"github.com/spf13/cobra"
)

func main() {
	println("ServerURL: ", cmd.ServerURL)
	if cmd.ServerURL == "" {
		panic("The env cmd.ServerURL is required")
	}

	rootCmd := buildCommands()
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		os.Exit(1)
	}
}

func buildCommands() *cobra.Command {
	userHomeDir := api.UserHomeDir()
	ritchieHomeDir := api.RitchieHomeDir()

	//deps
	sessionManager := session.NewManager(ritchieHomeDir)
	workspaceManager := workspace.NewChecker(ritchieHomeDir)
	ctxFinder := rcontext.NewFinder(ritchieHomeDir)
	ctxSetter := rcontext.NewSetter(ritchieHomeDir, ctxFinder)
	ctxRemover := rcontext.NewRemover(ritchieHomeDir, ctxFinder)
	ctxFindSetter := rcontext.NewFindSetter(ritchieHomeDir, ctxFinder, ctxSetter)
	ctxFindRemover := rcontext.NewFindRemover(ritchieHomeDir, ctxFinder, ctxRemover)
	repoManager := formula.NewTeamRepoManager(ritchieHomeDir, cmd.ServerURL, http.DefaultClient, sessionManager)
	sessionValidator := sessteam.NewValidator(sessionManager)
	loginManager := secteam.NewLoginManager(ritchieHomeDir, cmd.ServerURL, security.OAuthProvider, http.DefaultClient, sessionManager)
	logoutManager := secteam.NewLogoutManager(security.OAuthProvider, sessionManager, cmd.ServerURL)
	userManager := secteam.NewUserManager(cmd.ServerURL, http.DefaultClient, sessionManager)
	credSetter := credteam.NewSetter(cmd.ServerURL, http.DefaultClient, sessionManager, ctxFinder)
	credFinder := credteam.NewFinder(cmd.ServerURL, http.DefaultClient, sessionManager, ctxFinder)
	credSettings := credteam.NewSettings(cmd.ServerURL, http.DefaultClient, sessionManager, ctxFinder)
	treeManager := formula.NewTreeManager(ritchieHomeDir, repoManager, api.TeamCoreCmds)
	autocompleteGen := autocomplete.NewGenerator(treeManager)
	credResolver := envcredential.NewResolver(credFinder)
	envResolvers := make(env.Resolvers)
	envResolvers[env.Credential] = credResolver

	formulaRunner := formula.NewRunner(ritchieHomeDir, envResolvers, http.DefaultClient, treeManager)
	formulaCreator := formula.NewCreator(userHomeDir, treeManager)

	//commands
	rootCmd := cmd.NewRootCmd(workspaceManager, loginManager, repoManager, sessionValidator, api.Team)

	// level 1
	autocompleteCmd := cmd.NewAutocompleteCmd()
	addCmd := cmd.NewAddCmd()
	cleanCmd := cmd.NewCleanCmd()
	createCmd := cmd.NewCreateCmd()
	deleteCmd := cmd.NewDeleteCmd()
	listCmd := cmd.NewListCmd()
	loginCmd := cmd.NewLoginCmd(loginManager, repoManager)
	logoutCmd := cmd.NewLogoutCmd(logoutManager)
	setCmd := cmd.NewSetCmd()
	showCmd := cmd.NewShowCmd()
	updateCmd := cmd.NewUpdateCmd()

	// level 2
	setCredentialCmd := cmd.NewTeamSetCredentialCmd(credSetter, credSettings)
	createUserCmd := cmd.NewCreateUserCmd(userManager)
	deleteUserCmd := cmd.NewDeleteUserCmd(userManager)
	deleteCtxCmd := cmd.NewDeleteContextCmd(ctxFindRemover)
	setCtxCmd := cmd.NewSetContextCmd(ctxFindSetter)
	showCtxCmd := cmd.NewShowContextCmd(ctxFinder)
	addRepoCmd := cmd.NewAddRepoCmd(repoManager)
	cleanRepoCmd := cmd.NewCleanRepoCmd(repoManager)
	deleteRepoCmd := cmd.NewDeleteRepoCmd(repoManager)
	listRepoCmd := cmd.NewListRepoCmd(repoManager)
	updateRepoCmd := cmd.NewUpdateRepoCmd(repoManager)
	autocompleteZsh := cmd.NewAutocompleteZsh(autocompleteGen)
	autocompleteBash := cmd.NewAutocompleteBash(autocompleteGen)
	createFormulaCmd := cmd.NewCreateFormulaCmd(formulaCreator)

	autocompleteCmd.AddCommand(autocompleteZsh, autocompleteBash)
	addCmd.AddCommand(addRepoCmd)
	cleanCmd.AddCommand(cleanRepoCmd)
	createCmd.AddCommand(createUserCmd, createFormulaCmd)
	deleteCmd.AddCommand(deleteUserCmd, deleteRepoCmd, deleteCtxCmd)
	listCmd.AddCommand(listRepoCmd)
	setCmd.AddCommand(setCredentialCmd, setCtxCmd)
	showCmd.AddCommand(showCtxCmd)
	updateCmd.AddCommand(updateRepoCmd)

	rootCmd.AddCommand(addCmd, autocompleteCmd, cleanCmd, createCmd, deleteCmd, listCmd, loginCmd, logoutCmd, setCmd, showCmd, updateCmd)

	formulaCmd := cmd.NewFormulaCommand(api.TeamCoreCmds, treeManager, formulaRunner)
	if err := formulaCmd.Add(rootCmd); err != nil {
		panic(err)
	}

	sendMetrics(sessionManager)

	return rootCmd
}

func sendMetrics(sm session.DefaultManager) {
		metricsManager := metrics.NewSender(cmd.ServerURL, &http.Client{Timeout: 2 * time.Second}, sm)
		go metricsManager.SendCommand()
}
