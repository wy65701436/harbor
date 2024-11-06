/* tslint:disable */
export enum ArtifactTreeItemType {
  FILE = 'file',
  FOLDER = 'folder'
}

/**
* The artifact file tree item object.
*/
export interface ModelFileOverview {
  file?: string;
  file_hash?: string;
}

/**
* The artifact file tree item object.
*/
export interface ArtifactTreeItem {
  /**
   * The file or folder name.
   */
  name?: string;

  /**
   * The type of the artifact tree item.
   */
  file_type?: ArtifactTreeItemType;

  file_hash?: string;

  /** The original file name returned from backend. */
  file?: string;

  /**
   * The end time of the scan process that generating report
   */
  children?: ArtifactTreeItem[];

  /** The parent folder item */
  parent?: ArtifactTreeItem;

  /**
   * Is root folder or not. The root folder should not be displayed in UI.
   */
  isRootFolder?: boolean;
  isBackwardFolder?: boolean;
}
export const FOLDER_PATH_SEPARATOR: string = '/';

export function searchArtifactTreeItemByNameAndType(
  list: ArtifactTreeItem[],
  fileName: string,
  fileType: ArtifactTreeItemType
): ArtifactTreeItem[] {
  return list.filter(item => {
      return item.name == fileName && item.file_type == fileType;
  }) ?? [];
}

export function sortFunc(a: ArtifactTreeItem, b: ArtifactTreeItem): number {
  return a.name > b.name ? 1 : a.name === b.name ? 0 : -1;
}

export function sortArtifactTreeItems(list: ArtifactTreeItem[]): ArtifactTreeItem[] {
  const backwardArtifactTreeItems = list.filter(item => item.isBackwardFolder);
  const folderArtifactTreeItems = list.filter(
      item => !item.isBackwardFolder && item.file_type === ArtifactTreeItemType.FOLDER
  ).sort(sortFunc);
  const fileArtifactTreeItems = list.filter(
      item => item.file_type === ArtifactTreeItemType.FILE
  ).sort(sortFunc);
  return [...backwardArtifactTreeItems, ...folderArtifactTreeItems, ...fileArtifactTreeItems];
}

export function createArtifactTreeItem(
  root: ArtifactTreeItem,
  name: string,
  file_hash: string,
  isBackwardFolder: boolean = false,
  file_type: ArtifactTreeItemType = ArtifactTreeItemType.FOLDER,
  children: ArtifactTreeItem[] = [],
  isRootFolder: boolean = false
): ArtifactTreeItem {
  return {
      parent: root,
      name: name,
      file_hash: file_hash,
      isBackwardFolder: isBackwardFolder,
      file_type: file_type,
      children: children,
      isRootFolder: isRootFolder
  }
}

export function createArtifactTreeItems(fileName: string, fileHash: string, root: ArtifactTreeItem) {
  if (fileName.includes(FOLDER_PATH_SEPARATOR)) {
      const index = fileName.indexOf(FOLDER_PATH_SEPARATOR);
      const folerName = fileName.substring(0, index);
      const leftString = fileName.substring(index + 1);
      // check folerName object is exist or not.
      let matchedItems = searchArtifactTreeItemByNameAndType(root.children, folerName, ArtifactTreeItemType.FOLDER);
      if (matchedItems.length === 0) {
          const newArtifactTreeItemFolder = createArtifactTreeItem(
              root,
              folerName,
              "",
              false,
              ArtifactTreeItemType.FOLDER,
              [createArtifactTreeItem(root, '...', '', true)]
          );
          root.children.push(newArtifactTreeItemFolder);
          matchedItems.push(newArtifactTreeItemFolder);
      }
      createArtifactTreeItems(leftString, fileHash, matchedItems[0]);
  } else {
      const newArtifactTreeItem = createArtifactTreeItem(
          root,
          fileName,
          fileHash,
          false,
          ArtifactTreeItemType.FILE);
      root.children.push(newArtifactTreeItem);
  }
}

export function initRootArtifactTreeItem(modelFileOverviews: ModelFileOverview[], root: ArtifactTreeItem) {
  modelFileOverviews.forEach(item => {
      createArtifactTreeItems(item.file, item.file_hash, root);
  })
}