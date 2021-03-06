package storage_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	qstsv1a1 "code.cloudfoundry.org/quarks-statefulset/pkg/kube/apis/quarksstatefulset/v1alpha1"
	"code.cloudfoundry.org/quarks-utils/testing/machine"
	helper "code.cloudfoundry.org/quarks-utils/testing/testhelper"
)

var _ = Describe("QuarksStatefulSet", func() {
	var (
		quarksStatefulSet qstsv1a1.QuarksStatefulSet
	)

	BeforeEach(func() {
		essName := fmt.Sprintf("testess-%s", helper.RandString(5))
		quarksStatefulSet = env.DefaultQuarksStatefulSet(essName)

	})

	AfterEach(func() {
		Expect(env.WaitForPodsDelete(env.Namespace)).To(Succeed())
		// Skipping wait for PVCs to be deleted until the following is fixed
		// https://www.pivotaltracker.com/story/show/166896791
		// Expect(env.WaitForPVCsDelete(env.Namespace)).To(Succeed())
	})

	Context("when volumeClaimTemplates are specified", func() {
		BeforeEach(func() {

			essName := fmt.Sprintf("testess-%s", helper.RandString(5))
			storageClass, ok := os.LookupEnv("OPERATOR_TEST_STORAGE_CLASS")
			Expect(ok).To(BeTrue())
			quarksStatefulSet = env.QuarksStatefulSetWithPVC(essName, "pvc", storageClass)

		})

		It("should access same volume from different versions at the same time", func() {

			By("Adding volume write command to pod spec template")
			quarksStatefulSet.Spec.Template.Spec.Template.Spec.Containers[0].Image = "opensuse/leap:15.1"
			quarksStatefulSet.Spec.Template.Spec.Template.Spec.Containers[0].Command = []string{"/bin/bash", "-c", "echo present > /etc/random/presentFile ; sleep 3600"}

			By("Creating an QuarksStatefulSet")
			ess, tearDown, err := env.CreateQuarksStatefulSet(env.Namespace, quarksStatefulSet)
			Expect(err).NotTo(HaveOccurred())
			Expect(ess).NotTo(Equal(nil))
			defer func(tdf machine.TearDownFunc) { Expect(tdf()).To(Succeed()) }(tearDown)

			By("Checking for pod")
			err = env.WaitForPods(env.Namespace, "testpod=yes")
			Expect(err).NotTo(HaveOccurred())

			ess, err = env.GetQuarksStatefulSet(env.Namespace, ess.GetName())
			Expect(err).NotTo(HaveOccurred())
			Expect(ess).NotTo(Equal(nil))

			By("Updating the QuarksStatefulSet")
			ess.Spec.Template.Spec.Template.Spec.Containers[0].Command = []string{"/bin/bash", "-c", "cat /etc/random/presentFile ; sleep 3600"}
			ess.Spec.Template.Spec.Template.ObjectMeta.Labels["testpodupdated"] = "yes"
			essUpdated, tearDown, err := env.UpdateQuarksStatefulSet(env.Namespace, *ess)
			Expect(err).NotTo(HaveOccurred())
			Expect(essUpdated).NotTo(Equal(nil))
			defer func(tdf machine.TearDownFunc) { Expect(tdf()).To(Succeed()) }(tearDown)

			By("Checking for pod")
			err = env.WaitForPods(env.Namespace, "testpodupdated=yes")
			Expect(err).NotTo(HaveOccurred())

			podName := fmt.Sprintf("%s-%d", ess.GetName(), 0)

			out, err := env.GetPodLogs(env.Namespace, podName)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(Equal("present\n"))
		})

		It("should warn when volume templates are updated in qsts", func() {
			ess, tearDown, err := env.CreateQuarksStatefulSet(env.Namespace, quarksStatefulSet)
			Expect(err).NotTo(HaveOccurred())
			Expect(ess).NotTo(Equal(nil))
			defer func(tdf machine.TearDownFunc) { Expect(tdf()).To(Succeed()) }(tearDown)

			By("Checking for pod")
			err = env.WaitForPods(env.Namespace, "testpod=yes")
			Expect(err).NotTo(HaveOccurred())

			ess, err = env.GetQuarksStatefulSet(env.Namespace, ess.GetName())
			Expect(err).NotTo(HaveOccurred())
			Expect(ess).NotTo(Equal(nil))
			ess.Spec.Template.Spec.Template.ObjectMeta.Labels["testpodupdated"] = "yes"
			vols := []corev1.Volume{{
				Name: "pvc",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			}}
			ess.Spec.Template.Spec.Template.Spec.Volumes = vols
			ess.Spec.Template.Spec.VolumeClaimTemplates = nil
			By("Updating the QuarksStatefulSet")

			essUpdated, tearDown, err := env.UpdateQuarksStatefulSet(env.Namespace, *ess)
			Expect(err).NotTo(HaveOccurred())
			Expect(essUpdated).NotTo(Equal(nil))
			defer func(tdf machine.TearDownFunc) { Expect(tdf()).To(Succeed()) }(tearDown)
			By("Checking for sts")
			err = env.WaitForPods(env.Namespace, "testpodupdated=yes")
			Expect(err).NotTo(HaveOccurred())

			ess, err = env.GetQuarksStatefulSet(env.Namespace, ess.GetName())
			Expect(err).NotTo(HaveOccurred())
			Expect(ess).NotTo(Equal(nil))

			objectName := ess.ObjectMeta.Name
			objectUID := string(ess.ObjectMeta.UID)
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					objectName,
					objectUID,
					"VolumeClaimTemplatesWarning",
					"Change in VolumeClaimTemplates QuarksStatefulSet won't be performed in sts as it's not supported by Kubernetes",
				)
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
