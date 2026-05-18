<script lang="ts">
  import type {
    FieldMappingState,
    FieldRequirement,
    FieldRequirementGroup,
    MappingProfile,
    MappingSuggestion,
    PreviewRow,
    SchemaField,
    TargetRequirement
  } from "../types";
  import {
    canonicalColumn,
    cleanMapping,
    mappingHasValue,
    mappingInputValue,
    mappingModeLabel,
    normalizeMapping,
    resolveMappedValue
  } from "../mapping";
  import { readProfiles, saveProfile as persistProfile } from "../storage";

  type Props = {
    columns: string[];
    fields: SchemaField[];
    mapping: FieldMappingState;
    previewRows: PreviewRow[];
    sourcePlatform: string;
    suggestions: MappingSuggestion[];
    targetRequirements: TargetRequirement[];
    selectedTarget: string;
    onTargetChange: (target: string) => void;
    onMappingChange: (mapping: FieldMappingState) => void;
    onBack: () => void;
    onContinue: () => void;
  };

  let {
    columns,
    fields,
    mapping,
    previewRows,
    sourcePlatform,
    suggestions,
    targetRequirements,
    selectedTarget,
    onTargetChange,
    onMappingChange,
    onBack,
    onContinue
  }: Props = $props();
  let profileName = $state("Facebook Catalog Starter");
  let appliedSuggestionSignature = $state("");
  let savingProfile = $state(false);
  let saveError = $state("");
  let saveMessage = $state("");
  let profiles = $state<MappingProfile[]>([]);
  let selectedProfileId = $state("");
  let profileStatus = $state("");
  let profileLoadKey = $state("");

  let selectedRequirement = $derived(
    targetRequirements.find((requirement) => requirement.id === selectedTarget) ?? targetRequirements[0]
  );
  let requiredTotal = $derived(countRequiredItems(selectedRequirement));
  let requiredMapped = $derived(countMappedRequiredItems(selectedRequirement, mapping));
  let targetCompatibility = $derived(targetRequirements.map((requirement) => compatibilityForTarget(requirement)));
  let readyTargetCount = $derived(targetCompatibility.filter((item) => item.ready).length);
  let requiredSomewhereFields = $derived(requiredFieldsAcrossTargets(targetRequirements));
  let autoMappedCount = $derived(suggestions.filter((suggestion) => suggestion.source_column).length);
  let savedMapping = $derived(cleanMapping(mapping));
  let mappedFieldCount = $derived(Object.keys(savedMapping).length);
  let canSaveProfile = $derived(profileName.trim().length > 0 && mappedFieldCount > 0);

  $effect(() => {
    const signature = suggestions
      .map((suggestion) => `${suggestion.field}:${suggestion.source_column}`)
      .join("|");

    if (signature === appliedSuggestionSignature) {
      return;
    }

    const nextMapping: FieldMappingState = {};
    for (const suggestion of suggestions) {
      if (suggestion.source_column) {
        nextMapping[suggestion.field] = { mode: "column", value: suggestion.source_column };
      }
    }

    onMappingChange(nextMapping);
    appliedSuggestionSignature = signature;
  });

  $effect(() => {
    const key = `${sourcePlatform}:${selectedTarget}`;
    if (key === profileLoadKey) {
      return;
    }

    profileLoadKey = key;
    void loadProfiles();
  });

  function updateColumnMapping(fieldName: string, value: string) {
    const nextMapping = { ...mapping };
    if (!value) {
      delete nextMapping[fieldName];
    } else {
      nextMapping[fieldName] = { mode: "column", value };
    }

    onMappingChange(nextMapping);
    saveMessage = "";
    saveError = "";
  }

  function updateStaticMapping(fieldName: string, value: string) {
    const nextMapping = { ...mapping };
    if (!value.trim()) {
      delete nextMapping[fieldName];
    } else {
      nextMapping[fieldName] = { mode: "static", value: value.trim() };
    }

    onMappingChange(nextMapping);
    saveMessage = "";
    saveError = "";
  }

  function setStaticMode(fieldName: string, checked: boolean) {
    const nextMapping = { ...mapping };
    const currentValue = mapping[fieldName]?.value ?? "";

    if (checked) {
      nextMapping[fieldName] = { mode: "static", value: currentValue };
    } else {
      const column = canonicalColumn(currentValue, columns);
      if (column) {
        nextMapping[fieldName] = { mode: "column", value: column };
      } else {
        delete nextMapping[fieldName];
      }
    }

    onMappingChange(nextMapping);
    saveMessage = "";
    saveError = "";
  }

  function isStaticField(fieldName: string) {
    return mapping[fieldName]?.mode === "static";
  }

  function selectedColumn(fieldName: string) {
    const item = mapping[fieldName];
    return item?.mode === "column" ? item.value : "";
  }

  function canUseStaticValue(fieldName: string) {
    return isStaticField(fieldName) || !selectedColumn(fieldName);
  }

  async function loadProfiles() {
    const loadKey = `${sourcePlatform}:${selectedTarget}`;
    profileStatus = "Loading saved profiles";
    selectedProfileId = "";

    try {
      const savedProfiles = await readProfiles(sourcePlatform, selectedTarget);
      if (loadKey !== `${sourcePlatform}:${selectedTarget}`) {
        return;
      }

      profiles = savedProfiles;
      profileStatus = profiles.length > 0 ? `${profiles.length} saved profiles available` : "No saved profiles yet";
    } catch (err) {
      profiles = [];
      profileStatus = err instanceof Error ? err.message : "Could not load saved profiles";
    }
  }

  function applySelectedProfile() {
    const profile = profiles.find((item) => String(item.id) === selectedProfileId);
    if (!profile) {
      return;
    }

    onMappingChange(normalizeMapping(profile.mapping_json ?? {}, columns));
    profileName = profile.name;
    saveError = "";
    saveMessage = `Applied ${profile.name}`;
  }

  async function saveProfile() {
    if (!canSaveProfile || savingProfile) {
      return;
    }

    savingProfile = true;
    saveError = "";
    saveMessage = "";

    try {
      const profile = await persistProfile({
        name: profileName,
        source_platform: sourcePlatform,
        target_platform: selectedTarget,
        mapping_json: savedMapping
      });

      profiles = [profile, ...profiles.filter((item) => item.id !== profile.id)];
      selectedProfileId = String(profile.id);
      profileStatus = `${profiles.length} saved profiles available`;

      saveMessage = `Saved ${profile.name}`;
    } catch (err) {
      saveError = err instanceof Error ? err.message : "Could not save mapping profile";
    } finally {
      savingProfile = false;
    }
  }

  function countRequiredItems(requirement: TargetRequirement | undefined) {
    if (!requirement) {
      return 0;
    }

    return (
      requiredFields(requirement).length +
      requirementGroups(requirement).filter((group) => group.level === "required").length
    );
  }

  function countMappedRequiredItems(
    requirement: TargetRequirement | undefined,
    currentMapping: FieldMappingState
  ) {
    if (!requirement) {
      return 0;
    }

    const fieldCount = requiredFields(requirement).filter((item) => isFieldSatisfied(item.field, currentMapping)).length;
    const groupCount = requirementGroups(requirement).filter((group) => {
      if (group.level !== "required") {
        return false;
      }

      return isRequirementGroupSatisfied(group, currentMapping);
    }).length;

    return fieldCount + groupCount;
  }

  function compatibilityForTarget(requirement: TargetRequirement) {
    const missing: string[] = [];
    let satisfied = 0;
    const requiredItems = countRequiredItems(requirement);

    for (const item of requiredFields(requirement)) {
      if (isFieldSatisfied(item.field, mapping)) {
        satisfied += 1;
        continue;
      }

      missing.push(`${fieldLabel(item.field)}${item.field === "currency" ? " or derive from Price" : ""}`);
    }

    for (const group of requirementGroups(requirement).filter((item) => item.level === "required")) {
      if (isRequirementGroupSatisfied(group, mapping)) {
        satisfied += 1;
        continue;
      }

      missing.push(`At least ${group.min} of ${group.fields.map(fieldLabel).join(", ")}`);
    }

    return {
      requirement,
      ready: missing.length === 0,
      satisfied,
      requiredItems,
      missing
    };
  }

  function requirementForField(fieldName: string): FieldRequirement | null {
    if (!selectedRequirement) {
      return null;
    }

    const explicit = [
      ...requiredFields(selectedRequirement),
      ...conditionalFields(selectedRequirement),
      ...recommendedFields(selectedRequirement)
    ].find((requirement) => requirement.field === fieldName);

    if (explicit) {
      return explicit;
    }

    const group = requirementGroups(selectedRequirement).find((item) => item.fields.includes(fieldName));
    if (!group) {
      return null;
    }

    return {
      field: fieldName,
      level: group.level,
      note: group.note
    };
  }

  function requirementGroupForField(fieldName: string): FieldRequirementGroup | null {
    if (!selectedRequirement) {
      return null;
    }

    return requirementGroups(selectedRequirement).find((item) => item.fields.includes(fieldName)) ?? null;
  }

  function suggestionForField(fieldName: string) {
    return suggestions.find((suggestion) => suggestion.field === fieldName);
  }

  function isMissingRequired(
    fieldName: string,
    status: FieldRequirement | null,
    group: FieldRequirementGroup | null
  ) {
    if (status?.level !== "required" || isFieldSatisfied(fieldName, mapping)) {
      return false;
    }

    return group ? !isRequirementGroupSatisfied(group, mapping) : true;
  }

  function requirementBadge(status: FieldRequirement | null, group: FieldRequirementGroup | null) {
    if (!status) {
      return "";
    }
    if (isFieldDerived(status.field, mapping)) {
      return "derived";
    }
    if (group && status.level === "required") {
      return group.min === 1 ? "one required" : `${group.min} required`;
    }

    return status.level;
  }

  function groupStatus(group: FieldRequirementGroup | null) {
    if (!group) {
      return "";
    }

    const mappedCount = groupMappedCount(group, mapping);
    const labels = group.fields.map(fieldLabel).join(", ");
    if (mappedCount >= group.min) {
      return `Group satisfied: ${mappedCount} of ${labels} mapped.`;
    }

    return `Map at least ${group.min} of ${labels}.`;
  }

  function isRequirementGroupSatisfied(group: FieldRequirementGroup, currentMapping: FieldMappingState) {
    return groupMappedCount(group, currentMapping) >= group.min;
  }

  function groupMappedCount(group: FieldRequirementGroup, currentMapping: FieldMappingState) {
    return group.fields.filter((field) => isFieldSatisfied(field, currentMapping)).length;
  }

  function isFieldSatisfied(fieldName: string, currentMapping: FieldMappingState) {
    return mappingHasValue(currentMapping, fieldName) || isFieldDerived(fieldName, currentMapping);
  }

  function isFieldDerived(fieldName: string, currentMapping: FieldMappingState) {
    if (fieldName !== "currency" || !mappingHasValue(currentMapping, "price")) {
      return false;
    }

    return previewRows.some((row) => resolveMappedValue(currentMapping, "currency", row.values));
  }

  function fieldLabel(fieldName: string) {
    return fields.find((field) => field.name === fieldName)?.label ?? fieldName;
  }

  function requiredFieldsAcrossTargets(requirements: TargetRequirement[]) {
    const fieldsByName = new Map<string, number>();
    for (const requirement of requirements) {
      for (const item of requiredFields(requirement)) {
        fieldsByName.set(item.field, (fieldsByName.get(item.field) ?? 0) + 1);
      }
      for (const group of requirementGroups(requirement).filter((item) => item.level === "required")) {
        for (const field of group.fields) {
          fieldsByName.set(field, (fieldsByName.get(field) ?? 0) + 1);
        }
      }
    }

    return [...fieldsByName.entries()]
      .sort((a, b) => b[1] - a[1] || fieldLabel(a[0]).localeCompare(fieldLabel(b[0])))
      .map(([field, count]) => ({ field, label: fieldLabel(field), count }));
  }

  function requiredFields(requirement: TargetRequirement | undefined) {
    return requirement?.required_fields ?? [];
  }

  function conditionalFields(requirement: TargetRequirement | undefined) {
    return requirement?.conditional_fields ?? [];
  }

  function recommendedFields(requirement: TargetRequirement | undefined) {
    return requirement?.recommended_fields ?? [];
  }

  function requirementGroups(requirement: TargetRequirement | undefined) {
    return requirement?.requirement_groups ?? [];
  }
</script>

<section class="panel mapping-panel">
  <div class="panel-heading">
    <div>
      <h2>Field Mapping</h2>
      <p>
        {requiredMapped} of {requiredTotal} required items mapped for {selectedRequirement?.label ?? "target"} ·
        {autoMappedCount} auto-mapped
      </p>
    </div>
  </div>

  <form
    class="profile-bar"
    onsubmit={(event) => {
      event.preventDefault();
      saveProfile();
    }}
  >
    <label>
      Target
      <select value={selectedTarget} onchange={(event) => onTargetChange(event.currentTarget.value)}>
        {#each targetRequirements as target}
          <option value={target.id}>{target.label}</option>
        {/each}
      </select>
    </label>
    <label>
      Profile name
      <input aria-label="Mapping profile name" bind:value={profileName} />
    </label>
    <label>
      Saved profile
      <select bind:value={selectedProfileId} disabled={profiles.length === 0}>
        <option value="">{profiles.length === 0 ? "No profiles saved" : "Choose a profile"}</option>
        {#each profiles as profile}
          <option value={String(profile.id)}>{profile.name}</option>
        {/each}
      </select>
    </label>
    <div class="profile-apply">
      <button
        class="secondary"
        type="button"
        disabled={!selectedProfileId}
        onclick={applySelectedProfile}
      >
        Apply Profile
      </button>
      <small>{profileStatus}</small>
    </div>
    <div class="save-control">
      <button type="submit" disabled={!canSaveProfile || savingProfile}>
        {savingProfile ? "Saving..." : "Save Mapping Profile"}
      </button>
      <small class:error={Boolean(saveError)}>
        {#if saveError}
          {saveError}
        {:else if saveMessage}
          {saveMessage}
        {:else if canSaveProfile}
          Ready to save {mappedFieldCount} mapped fields
        {:else}
          Map at least one field to save
        {/if}
      </small>
    </div>
  </form>

  <div class="mapping-workspace">
    <div class="mapping-list">
      {#each fields as field}
        {@const status = requirementForField(field.name)}
        {@const group = requirementGroupForField(field.name)}
        {@const suggestion = suggestionForField(field.name)}
        <div class="field-card" class:missing-required={isMissingRequired(field.name, status, group)}>
        <div class="field-title">
          <span>{field.label}</span>
          {#if status}
            <em
              class={status.level}
              class:flexible={Boolean(group)}
              class:derived={isFieldDerived(field.name, mapping)}
            >
              {requirementBadge(status, group)}
            </em>
          {/if}
        </div>

        <div class="mapping-control">
          {#if isStaticField(field.name)}
            <label class="input-label">
              Static value
              <input
                value={mappingInputValue(mapping, field.name)}
                disabled={columns.length === 0}
                placeholder="Value for every row"
                oninput={(event) => updateStaticMapping(field.name, event.currentTarget.value)}
                onblur={(event) => updateStaticMapping(field.name, event.currentTarget.value)}
              />
            </label>
          {:else}
            <label class="input-label">
              Source column
              <select
                value={selectedColumn(field.name)}
                disabled={columns.length === 0}
                onchange={(event) => updateColumnMapping(field.name, event.currentTarget.value)}
              >
                <option value="">Not mapped</option>
                {#each columns as column}
                  <option value={column}>{column}</option>
                {/each}
              </select>
            </label>
          {/if}

          {#if canUseStaticValue(field.name)}
            <label class="static-toggle" title="Use one value for every row">
              <input
                type="checkbox"
                checked={isStaticField(field.name)}
                disabled={columns.length === 0}
                onchange={(event) => setStaticMode(field.name, event.currentTarget.checked)}
              />
              <span>Static</span>
            </label>
          {/if}
        </div>

        {#if suggestion}
          <small>
            Auto: {suggestion.source_column} · {Math.round(suggestion.confidence * 100)}%
          </small>
        {/if}
        {#if mappingModeLabel(mapping, field.name)}
          <small class="mode">{mappingModeLabel(mapping, field.name)}</small>
        {/if}
        {#if group}
          <small class:mode={isRequirementGroupSatisfied(group, mapping)}>{groupStatus(group)}</small>
        {/if}
        {#if isFieldDerived(field.name, mapping)}
          <small class="mode">Derived from mapped Price values.</small>
        {/if}
        {#if status?.note}
          <small class="note">{status.note}</small>
        {/if}
      </div>
      {/each}
    </div>

    <aside class="compatibility-panel" aria-label="Export compatibility">
      <div class="compatibility-heading">
        <div>
          <h3>Export Compatibility</h3>
          <p>{readyTargetCount} of {targetRequirements.length} targets ready</p>
        </div>
      </div>

      <div class="target-list">
        {#each targetCompatibility as item}
          <button
            type="button"
            class:active={selectedTarget === item.requirement.id}
            class:ready={item.ready}
            onclick={() => onTargetChange(item.requirement.id)}
          >
            <span class="target-status">{item.ready ? "✓" : "×"}</span>
            <span>
              <strong>{item.requirement.label}</strong>
              <small>{item.satisfied} of {item.requiredItems} requirements met</small>
            </span>
          </button>
          {#if !item.ready}
            <ul>
              {#each item.missing.slice(0, 3) as missing}
                <li>{missing}</li>
              {/each}
              {#if item.missing.length > 3}
                <li>{item.missing.length - 3} more missing</li>
              {/if}
            </ul>
          {/if}
        {/each}
      </div>

      <div class="required-somewhere">
        <strong>Required by at least one format</strong>
        <div>
          {#each requiredSomewhereFields as item}
            <span class:met={isFieldSatisfied(item.field, mapping)}>
              {item.label}
              <small>{item.count}</small>
            </span>
          {/each}
        </div>
      </div>
    </aside>
  </div>

  <div class="actions">
    <button class="secondary" type="button" onclick={onBack}>Back to Upload</button>
    <button type="button" disabled={columns.length === 0} onclick={onContinue}>Continue to Validation</button>
  </div>
</section>

<style>
  .panel {
    display: grid;
    gap: 18px;
    border: 1px solid rgb(255 255 255 / 0.62);
    border-radius: 24px;
    background: rgb(255 255 255 / 0.66);
    box-shadow: 0 24px 70px rgb(31 42 55 / 0.12);
    backdrop-filter: blur(22px);
    padding: 20px;
  }

  .panel-heading {
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
    align-items: flex-start;
    justify-content: space-between;
    min-width: 0;
  }

  .panel-heading > div {
    flex: 1 1 280px;
    min-width: 0;
  }

  h2,
  p {
    margin: 0;
  }

  h2 {
    color: #13251c;
    font-size: 1.12rem;
  }

  p {
    max-width: 64ch;
    color: #62716a;
    line-height: 1.45;
    margin-top: 4px;
    overflow-wrap: anywhere;
  }

  .profile-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    align-items: end;
    border: 1px solid rgb(214 226 222 / 0.7);
    border-radius: 18px;
    background: rgb(255 255 255 / 0.46);
    padding: 14px;
  }

  .profile-bar label {
    flex: 1 1 220px;
    min-width: 0;
  }

  .profile-apply,
  .save-control {
    display: grid;
    flex: 1 1 250px;
    gap: 7px;
    min-width: min(100%, 250px);
  }

  .mapping-workspace {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(280px, 340px);
    gap: 16px;
    align-items: start;
  }

  .mapping-list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 340px), 1fr));
    gap: 12px;
  }

  .compatibility-panel {
    position: sticky;
    top: 16px;
    display: grid;
    gap: 14px;
    border: 1px solid rgb(214 226 222 / 0.76);
    border-radius: 18px;
    background: rgb(255 255 255 / 0.52);
    padding: 14px;
  }

  .compatibility-heading h3,
  .compatibility-heading p {
    margin: 0;
  }

  .compatibility-heading h3 {
    color: #13251c;
    font-size: 0.95rem;
  }

  .compatibility-heading p {
    color: #62716a;
    font-size: 0.82rem;
    font-weight: 700;
    line-height: 1.35;
    margin-top: 3px;
  }

  .target-list {
    display: grid;
    gap: 8px;
  }

  .target-list button {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    align-items: center;
    gap: 9px;
    min-height: 48px;
    border: 1px solid rgb(214 226 222 / 0.76);
    border-radius: 14px;
    background: rgb(255 255 255 / 0.58);
    box-shadow: none;
    color: #31463d;
    padding: 8px 10px;
    text-align: left;
  }

  .target-list button:hover,
  .target-list button.active {
    border-color: rgb(37 88 64 / 0.42);
    background: rgb(255 255 255 / 0.82);
    box-shadow: none;
    transform: none;
  }

  .target-list button.ready {
    border-color: rgb(37 88 64 / 0.22);
  }

  .target-status {
    display: inline-grid;
    width: 24px;
    height: 24px;
    place-items: center;
    border-radius: 999px;
    background: #fff1ed;
    color: #9b3324;
    font-size: 0.85rem;
    font-weight: 900;
  }

  .target-list button.ready .target-status {
    background: #e8f5ee;
    color: #255840;
  }

  .target-list button > span:last-child {
    display: grid;
    gap: 2px;
    min-width: 0;
  }

  .target-list strong {
    color: #13251c;
    font-size: 0.82rem;
    overflow-wrap: anywhere;
  }

  .target-list ul {
    display: grid;
    gap: 4px;
    margin: -2px 0 4px 33px;
    padding-left: 16px;
  }

  .target-list li {
    color: #7c3f0f;
    font-size: 0.76rem;
    font-weight: 700;
    line-height: 1.35;
  }

  .required-somewhere {
    display: grid;
    gap: 9px;
    border-top: 1px solid rgb(218 228 224 / 0.78);
    padding-top: 12px;
  }

  .required-somewhere > strong {
    color: #13251c;
    font-size: 0.82rem;
  }

  .required-somewhere div {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .required-somewhere span {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    border: 1px solid rgb(217 119 6 / 0.24);
    border-radius: 999px;
    background: rgb(255 251 235 / 0.72);
    color: #7c3f0f;
    font-size: 0.74rem;
    font-weight: 800;
    padding: 4px 7px;
  }

  .required-somewhere span.met {
    border-color: rgb(37 88 64 / 0.18);
    background: #eef8f2;
    color: #255840;
  }

  .required-somewhere small {
    color: inherit;
    font-size: 0.68rem;
    opacity: 0.78;
  }

  label {
    display: grid;
    gap: 7px;
    color: #33443c;
    font-weight: 700;
    min-width: 0;
  }

  .field-card {
    display: grid;
    gap: 10px;
    border: 1px solid rgb(214 226 222 / 0.76);
    border-radius: 18px;
    background: rgb(255 255 255 / 0.52);
    padding: 13px;
  }

  .field-card.missing-required {
    border-color: rgb(220 38 38 / 0.78);
    box-shadow: inset 4px 0 0 rgb(220 38 38 / 0.78);
  }

  .field-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    min-width: 0;
  }

  .field-title > span {
    min-width: 0;
    overflow-wrap: anywhere;
  }

  .mapping-control {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 10px;
    align-items: end;
  }

  .input-label {
    gap: 6px;
    font-size: 0.78rem;
  }

  .static-toggle {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-height: 42px;
    border: 1px solid rgb(194 210 204 / 0.85);
    border-radius: 12px;
    background: rgb(255 255 255 / 0.58);
    color: #40564d;
    font-size: 0.76rem;
    font-weight: 800;
    padding: 0 9px;
    white-space: nowrap;
  }

  .static-toggle input {
    width: 14px;
    min-width: 14px;
    min-height: 14px;
    accent-color: #1d563b;
    padding: 0;
  }

  em {
    color: #8a4a18;
    font-size: 0.72rem;
    font-style: normal;
    font-weight: 800;
    line-height: 1;
    text-transform: uppercase;
    white-space: nowrap;
  }

  em.required {
    color: #9b3324;
  }

  em.conditional {
    color: #8a4a18;
  }

  em.recommended {
    color: #255840;
  }

  em.flexible {
    color: #255840;
  }

  em.derived {
    color: #255840;
  }

  input,
  select {
    width: 100%;
    min-width: 0;
    min-height: 42px;
    border: 1px solid rgb(194 210 204 / 0.95);
    border-radius: 12px;
    background: rgb(255 255 255 / 0.82);
    color: #13251c;
    padding: 0 12px;
  }

  .field-card select {
    text-overflow: ellipsis;
  }

  small {
    color: #5c7168;
    font-size: 0.78rem;
    font-weight: 700;
    line-height: 1.35;
    overflow-wrap: anywhere;
  }

  small.error {
    color: #9b3324;
  }

  small.mode {
    color: #255840;
  }

  small.note {
    color: #708078;
    font-weight: 600;
    line-height: 1.35;
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: 12px;
    border-top: 1px solid rgb(218 228 224 / 0.78);
    padding-top: 16px;
  }

  .actions button {
    min-width: 180px;
  }

  .secondary {
    border: 1px solid rgb(194 210 204 / 0.9);
    background: rgb(255 255 255 / 0.72);
    color: #244f3b;
  }

  @media (max-width: 820px) {
    .panel {
      border-radius: 18px;
    }

    .mapping-workspace {
      grid-template-columns: 1fr;
    }

    .compatibility-panel {
      position: static;
    }

    .actions button {
      width: 100%;
    }
  }
</style>
