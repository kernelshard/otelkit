# GitHub Release Permissions Guide

## Fixing 403 Permission Errors

The GitHub Actions workflow is encountering 403 permission errors when trying to create releases. Here's how to fix this:

### Step 1: Check Repository Settings

1. Go to your GitHub repository: https://github.com/samims/otelkit
2. Navigate to **Settings → Actions → General**
3. Under **Workflow permissions**, ensure:
   - ✅ **Read and write permissions** is selected
   - ✅ **Allow GitHub Actions to create and approve pull requests** is checked

### Step 2: Alternative Solution - Personal Access Token

If the above doesn't work, create a personal access token:

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token" → "Generate new token (classic)"
3. Set these permissions:
   - `repo` (full control of private repositories)
   - `workflow`
4. Copy the generated token
5. Go to your repository Settings → Secrets and variables → Actions
6. Add a new secret:
   - Name: `PERSONAL_ACCESS_TOKEN`
   - Value: [paste your token here]

### Step 3: Update Workflow (If Using PAT)

If you create a personal access token, update the workflow to use it:

```yaml
- name: Create release
  uses: softprops/action-gh-release@v1
  with:
    files: |
      dist/README.txt
    generate_release_notes: true
  env:
    GITHUB_TOKEN: ${{ secrets.PERSONAL_ACCESS_TOKEN }}
```

### Step 4: Enable GitHub Pages

For documentation deployment:

1. Go to repository Settings → Pages
2. Under **Build and deployment**, select:
   - **Source**: GitHub Actions
3. Save the settings

## Current Workflow Status

The workflow has been updated to:
- ✅ Remove binary building (since this is a library)
- ✅ Simplify documentation generation
- ✅ Remove Docker build (not needed for library)
- ✅ Use proper file patterns for releases

Once permissions are fixed, the workflow should successfully:
1. Run all tests
2. Create a GitHub Release
3. Generate and deploy API documentation

## Testing the Fix

After updating permissions, you can trigger the workflow by:
```bash
git tag -d v0.1.0
git tag v0.1.0
git push origin --force v0.1.0
```

Check the Actions tab in your repository to monitor progress.
