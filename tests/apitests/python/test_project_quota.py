from __future__ import absolute_import
import unittest

from testutils import harbor_server
from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_image_to_project
from library.system import System

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        project = Project()
        self.project= project

        user = User()
        self.user= user

        repo = Repository()
        self.repo= repo

        self.system = System()

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == True, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(TestProjects.repo_name, **TestProjects.USER_SCAN_IMAGE_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_test_quota_id, **TestProjects.USER_SCAN_IMAGE_CLIENT)

        #3. Delete user(UA);
        self.user.delete_user(TestProjects.user_test_quota_id, **ADMIN_CLIENT)

    def testProjectQuota(self):
        """
        Test case:
            Project Quota
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new private project(PA) by user(UA);
            3. Add user(UA) as a member of project(PA) with project-admin role;
            4. Push an image to project(PA) by user(UA), then check the project quota usage;
            5. Update the project storage quota to 200 MB;
            6. Push another image which size is larger than 200MB, then check the client error message and project quota usage.
        Tear down:
            1. Delete repository(RA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA);
        """
        url = ADMIN_CLIENT["endpoint"]
        user_001_password = "Aa123456"

        #1. Create user-001
        TestProjects.user_test_quota_id, user_test_quota_name = self.user.create_user(user_password = user_001_password, **ADMIN_CLIENT)
        TestProjects.USER_TEST_QUOTA_CLIENT=dict(endpoint = url, username = user_test_quota_name, password = user_001_password)

        #2. Create a new private project(PA) by user(UA);
        TestProjects.project_test_quota_id, project_test_quota_name = self.project.create_project(metadata = {"public": "false"}, **ADMIN_CLIENT)

        #3. Add user(UA) as a member of project(PA) with project-admin role;
        self.project.add_project_members(TestProjects.project_test_quota_id, TestProjects.user_test_quota_id, **ADMIN_CLIENT)

        #4.Push an image to project(PA) by user(UA), then check the project quota usage; -- {"count": 1, "storage": 2791709}
        image = "alpine"
        src_tag = "3.10"
        TestProjects.repo_name, tag = push_image_to_project(project_test_quota_name, harbor_server, user_test_quota_name, user_001_password, image, src_tag)

        quota = self.system.get_project_quota("project", TestProjects.project_test_quota_id, **ADMIN_CLIENT)
        self.assertEqual(quota[0].used["count"], 1)
        self.assertEqual(quota[0].used["storage"], 2791709)

        #5. Update the project storage quota to 200 MB;
        self.system.set_project_quota(TestProjects.project_test_quota_id, "{\"hard\":{\"count\":10,\"storage\":209715200}}", **ADMIN_CLIENT)

        #6. Push another image which size is larger than 200MB, then check the client error message and project quota usage.
        image = "node"
        src_tag = "12.10.0"
        TestProjects.repo_name, tag = push_image_to_project(project_test_quota_name, harbor_server, user_test_quota_name, user_001_password, image, src_tag)
        


if __name__ == '__main__':
    unittest.main()